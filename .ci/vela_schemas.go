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
		set -euo pipefail

		# ── Helper: dump tailscaled log and exit with a clear message ─────────────
		die() {
			echo "" >&2
			echo "╔══════════════════════════════════════════════════════╗" >&2
			echo "║  FATAL: $1" >&2
			echo "╚══════════════════════════════════════════════════════╝" >&2
			echo "" >&2
			echo "▼▼▼ tailscaled daemon log ▼▼▼" >&2
			cat /var/log/tailscaled.log >&2 || echo "(log not available)" >&2
			echo "▲▲▲ end of tailscaled log ▲▲▲" >&2
			exit 1
		}

		# ── PHASE 1: Start Tailscale daemon ───────────────────────────────────────
		echo "━━━ PHASE 1: Starting Tailscale daemon ━━━" >&2
		mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale /var/log
		tailscaled --tun=userspace-networking --socks5-server=localhost:1055 \
			--statedir=/var/lib/tailscale > /var/log/tailscaled.log 2>&1 &
		TAILSCALED_PID=$!

		echo "  Waiting 10s for daemon to initialize (PID $TAILSCALED_PID)..." >&2
		sleep 10

		# Verify the daemon process is still alive
		if ! kill -0 $TAILSCALED_PID 2>/dev/null; then
			die "tailscaled exited immediately after launch"
		fi
		echo "  ✅ Daemon is running." >&2

		# ── PHASE 2: Authenticate ─────────────────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 2: Authenticating with Tailscale ━━━" >&2
		tailscale up --authkey="$TAILSCALE_AUTH_KEY" --hostname=ci-nucleus-runner \
			--timeout=30s >&2 || die "tailscale up failed (bad auth key or timeout)"
		echo "  ✅ Authentication succeeded." >&2

		# ── PHASE 3: Verify Tailscale peer status ─────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 3: Checking Tailscale peer status ━━━" >&2
		tailscale status >&2
		echo "" >&2

		# Extract the cluster IP from what we will receive in the kubeconfig
		# so we can ping it before wasting time on kubectl
		echo "$KUBECONFIG_BASE64" | base64 -d > /tmp/kubeconfig
		sed -i 's/certificate-authority-data:.*/insecure-skip-tls-verify: true/g' /tmp/kubeconfig
		export KUBECONFIG=/tmp/kubeconfig

		CLUSTER_IP=$(grep "server:" /tmp/kubeconfig | awk '{print $2}' | head -1 | sed 's|https\?://||' | cut -d: -f1)
		CLUSTER_PORT=$(grep "server:" /tmp/kubeconfig | awk '{print $2}' | head -1 | sed 's|https\?://||' | cut -d: -f2)
		echo "  Target cluster  : $CLUSTER_IP:$CLUSTER_PORT" >&2

		# ── PHASE 4: Ping the cluster IP over Tailnet ─────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 4: Pinging cluster IP over Tailnet ━━━" >&2
		echo "  Running: tailscale ping --timeout=15s $CLUSTER_IP" >&2
		if tailscale ping --timeout=15s "$CLUSTER_IP" >&2; then
			echo "  ✅ Tailnet ping succeeded — node is reachable!" >&2
		else
			echo "  ⚠️  tailscale ping failed. Possible reasons:" >&2
			echo "     • The cluster node is not online in your Tailnet" >&2
			echo "     • The key was rejected / not yet propagated" >&2
			echo "     • MagicDNS needs a moment — will still try SOCKS5" >&2
		fi

		# Give routes a moment to propagate after auth
		echo "  Waiting 5s for route propagation..." >&2
		sleep 5

		# ── PHASE 5: Verify SOCKS5 proxy with curl (no kubectl yet) ──────────────
		echo "" >&2
		echo "━━━ PHASE 5: Testing SOCKS5 proxy connectivity ━━━" >&2
		echo "  Running: curl through socks5://localhost:1055 to cluster endpoint" >&2
		CURL_OUT=$(curl --silent --max-time 10 \
			--proxy socks5://localhost:1055 \
			--insecure \
			--write-out "HTTP_STATUS:%{http_code}" \
			"https://$CLUSTER_IP:$CLUSTER_PORT/healthz" 2>&1 || true)
		echo "  curl response: $CURL_OUT" >&2
		if echo "$CURL_OUT" | grep -qE "HTTP_STATUS:(200|401|403)"; then
			echo "  ✅ SOCKS5 proxy is routing to the cluster API server!" >&2
		else
			echo "  ⚠️  curl did not get a 200/401/403 back." >&2
			echo "     This means the SOCKS5 proxy cannot reach $CLUSTER_IP:$CLUSTER_PORT" >&2
			echo "     Likely cause: Tailscale userspace routing is not complete yet." >&2
			die "SOCKS5 proxy connectivity check failed — kubectl will not work"
		fi

		# ── PHASE 6: kubectl connectivity check ───────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 6: Routing kubectl through SOCKS5 proxy ━━━" >&2
		export ALL_PROXY=socks5://localhost:1055

		MAX_ATTEMPTS=5
		for i in $(seq 1 $MAX_ATTEMPTS); do
			echo "  Attempt $i/$MAX_ATTEMPTS — kubectl cluster-info..." >&2
			if kubectl cluster-info 2>&1 | tee /tmp/kubectl_out.txt >&2; then
				echo "  ✅ Cluster is reachable!" >&2
				break
			fi
			echo "  ✖ kubectl failed. Last error:" >&2
			cat /tmp/kubectl_out.txt >&2
			if [ "$i" -eq "$MAX_ATTEMPTS" ]; then
				die "kubectl could not reach the cluster after $MAX_ATTEMPTS attempts"
			fi
			echo "  Retrying in 10s..." >&2
			sleep 10
		done

		# ── PHASE 7: Fetch KubeVela schemas ───────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 7: Fetching KubeVela ConfigMaps ━━━" >&2
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
