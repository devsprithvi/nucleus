package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dagger/nucleus-ci/internal/dagger"
)

// SchemaMapList represents the expected JSON response from the Kubernetes API.
type SchemaMapList struct {
	Items []SchemaMapItem `json:"items"`
}

// SchemaMapItem represents a single Kubernetes ConfigMap containing a KubeVela schema.
type SchemaMapItem struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Data map[string]string `json:"data"`
}

// GenerateVelaSchemas extracts KubeVela schemas from the remote cluster and builds
// a master JSON schema for IDE auto-completion. It handles its own secure VPN connection.
func (a *Actions) GenerateVelaSchemas(
	ctx context.Context,
	tailscaleAuthKey *dagger.Secret,
	kubeconfigBase64 *dagger.Secret,
) (*dagger.Directory, error) {
	// Initialize the Dagger client for this execution context
	dag := dagger.Connect()

	// Phase 1: Establish the VPN connection and fetch the cluster data
	rawJSON, err := executeVPNAndFetch(ctx, dag, tailscaleAuthKey, kubeconfigBase64)
	if err != nil {
		return nil, fmt.Errorf("phase 1 (vpn connection & fetch) failed: %w", err)
	}

	// Phase 2: Parse the Kubernetes JSON response natively in Go
	var list SchemaMapList
	if err := json.Unmarshal([]byte(rawJSON), &list); err != nil {
		return nil, fmt.Errorf("phase 2 (parse json) failed: %w", err)
	}

	if len(list.Items) == 0 {
		return nil, fmt.Errorf("no schema configmaps found in the cluster")
	}

	// Phase 3: Create a new virtual Dagger Directory and write all individual schemas
	schemaDir := dag.Directory()
	schemaDir, extractedNames := writeIndividualSchemas(schemaDir, list)

	// Phase 4: Inject the Master Schema for IDE routing
	schemaDir, err = injectMasterSchema(schemaDir, extractedNames)
	if err != nil {
		return nil, fmt.Errorf("phase 4 (master schema routing) failed: %w", err)
	}

	// Return the final directory back to the pipeline caller
	return schemaDir, nil
}

// executeVPNAndFetch provisions an environment, establishes a Tailscale userspace
// network, proxies the connection, and safely extracts the KubeVela schemas.
func executeVPNAndFetch(
	ctx context.Context,
	dag *dagger.Client,
	tsKey *dagger.Secret,
	kubeconfigB64 *dagger.Secret,
) (string, error) {
	container := buildRunnerContainer(dag).
		WithSecretVariable("TAILSCALE_AUTH_KEY", tsKey).
		WithSecretVariable("KUBECONFIG_BASE64", kubeconfigB64)

	const script = `
		echo "Creating necessary directories for Tailscale..." >&2
		mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale /var/log
		
		echo "Starting Tailscale daemon in userspace mode..." >&2
		# CRITICAL FIX: We MUST redirect all output to a file so Dagger doesn't hang waiting for the stream to close.
		# Using --socks5-server to match the ALL_PROXY=socks5:// setting used by kubectl.
		tailscaled --tun=userspace-networking --socks5-server=localhost:1055 --statedir=/var/lib/tailscale > /var/log/tailscaled.log 2>&1 &
		
		echo "Waiting for daemon to initialize..." >&2
		sleep 5

		echo "Authenticating Tailscale..." >&2
		tailscale up --authkey="$TAILSCALE_AUTH_KEY" --hostname=ci-runner >&2

		echo "Checking Tailscale status..." >&2
		tailscale status >&2

		echo "Waiting for routes to propagate..." >&2
		sleep 3

		echo "Decoding Kubeconfig..." >&2
		echo "$KUBECONFIG_BASE64" | base64 -d > /tmp/kubeconfig
		
		# User requested override to avoid TLS verification errors over proxy
		sed -i 's/certificate-authority-data:.*/insecure-skip-tls-verify: true/g' /tmp/kubeconfig
		export KUBECONFIG=/tmp/kubeconfig

		# Diagnostic: show the target server URL so we can verify it's correct
		echo "Kubeconfig target server:" >&2
		grep "server:" /tmp/kubeconfig >&2

		echo "Routing kubectl through Tailscale SOCKS5 Proxy..." >&2
		export ALL_PROXY=socks5://localhost:1055

		echo "Waiting for Kube API to become reachable over Tailnet..." >&2
		for i in {1..6}; do
			if kubectl cluster-info >/dev/null 2>&1; then
				echo "Cluster is reachable!" >&2
				break
			fi
			echo "Waiting for cluster... attempt $i/6" >&2
			sleep 5
		done

		echo "Fetching KubeVela ConfigMaps..." >&2
		kubectl get configmap -n vela-system -l definition.oam.dev=schema -o json
	`

	return container.
		WithExec([]string{"bash", "-c", script}).
		Stdout(ctx)
}

// buildRunnerContainer abstracts the OS-level dependency installation, ensuring
// the primary functions remain focused solely on business logic.
func buildRunnerContainer(dag *dagger.Client) *dagger.Container {
	return dag.Container().
		// CHANGED: Upgraded to ubuntu:22.04 for better networking library support
		From("ubuntu:22.04").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "curl", "ca-certificates", "openssh-client"}).
		WithExec([]string{"bash", "-c", "curl -fsSL https://tailscale.com/install.sh | sh"}).
		WithExec([]string{"bash", "-c", "curl -LO https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && mv kubectl /usr/local/bin/"})
}

// writeIndividualSchemas loops through the retrieved ConfigMaps and writes each JSON schema into the Dagger directory.
func writeIndividualSchemas(dir *dagger.Directory, list SchemaMapList) (*dagger.Directory, []string) {
	var extractedNames []string

	for _, item := range list.Items {
		defName := strings.TrimPrefix(item.Metadata.Name, "schema-")
		schemaContent, ok := item.Data["openapi-v3-json-schema"]
		if !ok || schemaContent == "" {
			continue
		}

		extractedNames = append(extractedNames, defName)
		fileName := fmt.Sprintf("%s-schema.json", defName)

		// Insert the new file into the immutable virtual directory
		dir = dir.WithNewFile(fileName, schemaContent)
	}

	return dir, extractedNames
}

// injectMasterSchema dynamically builds the root vela-application-schema.json file to map 'type' to the correct schema file.
func injectMasterSchema(dir *dagger.Directory, componentNames []string) (*dagger.Directory, error) {
	var oneOfRules []map[string]interface{}

	for _, name := range componentNames {
		// Create the conditional routing logic for the IDE
		rule := map[string]interface{}{
			"if": map[string]interface{}{
				"properties": map[string]interface{}{
					"type": map[string]interface{}{"const": name},
				},
			},
			"then": map[string]interface{}{
				"properties": map[string]interface{}{
					"properties": map[string]interface{}{
						"$ref": fmt.Sprintf("%s-schema.json", name),
					},
				},
			},
		}
		oneOfRules = append(oneOfRules, rule)
	}

	// Construct the foundational KubeVela Application schema
	masterSchema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"title":   "KubeVela Application",
		"properties": map[string]interface{}{
			"apiVersion": map[string]interface{}{"const": "core.oam.dev/v1beta1"},
			"kind":       map[string]interface{}{"const": "Application"},
			"spec": map[string]interface{}{
				"properties": map[string]interface{}{
					"components": map[string]interface{}{
						"type": "array",
						"items": map[string]interface{}{
							"oneOf": oneOfRules,
						},
					},
				},
			},
		},
	}

	// Convert the map structure into a formatted JSON string
	masterJSONBytes, err := json.MarshalIndent(masterSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	// Add the stitched master schema to the directory
	return dir.WithNewFile("vela-application-schema.json", string(masterJSONBytes)), nil
}
