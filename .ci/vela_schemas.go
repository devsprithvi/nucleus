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
	var entries []schemaEntry
	schemaDir, entries = writeIndividualSchemas(schemaDir, list)

	// Phase 4: Inject the Master Schema for IDE routing
	schemaDir, err = injectMasterSchema(schemaDir, entries)
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
		# IMPORTANT: kubectl uses Go's net/http which only reads HTTPS_PROXY / HTTP_PROXY.
		# ALL_PROXY is a Unix convention that Go ignores — kubectl will dial directly without this.
		export HTTPS_PROXY=socks5://localhost:1055
		export HTTP_PROXY=socks5://localhost:1055
		# Unset ALL_PROXY to avoid confusion
		unset ALL_PROXY

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

// schemaEntry pairs the short KubeVela type name (what users write in YAML)
// with the full filename on disk and the definition category.
type schemaEntry struct {
	typeName string // short type users write, e.g. "webservice"
	fileName string // file on disk, e.g. "component-schema-webservice-schema.json"
	category string // "component" | "trait" | "policy" | "workflowstep"
}

// categoryInfo maps ConfigMap name prefixes to their definition category.
var categoryInfo = []struct {
	prefix   string
	category string
}{
	{"component-schema-", "component"},
	{"trait-schema-", "trait"},
	{"policy-schema-", "policy"},
	{"workflowstep-schema-", "workflowstep"},
}

// parseEntry derives the short type name and category from a KubeVela ConfigMap name.
// ConfigMap names follow: schema-<category-prefix><type-name>
// e.g. "schema-component-schema-webservice" → typeName="webservice", category="component"
// Returns empty strings if the entry should be skipped (e.g. versioned duplicates ending in -v<n>).
func parseEntry(configMapName string) (typeName, category string) {
	name := strings.TrimPrefix(configMapName, "schema-")
	for _, ci := range categoryInfo {
		if strings.HasPrefix(name, ci.prefix) {
			shortName := strings.TrimPrefix(name, ci.prefix)
			// Skip "-v1", "-v2" etc. – they are identical duplicates of the non-versioned entries.
			// A versioned entry ends in "-v" followed only by digits.
			if idx := strings.LastIndex(shortName, "-v"); idx >= 0 {
				suffix := shortName[idx+2:]
				isAllDigits := len(suffix) > 0
				for _, ch := range suffix {
					if ch < '0' || ch > '9' {
						isAllDigits = false
						break
					}
				}
				if isAllDigits {
					return "", "" // signal to skip
				}
			}
			return shortName, ci.category
		}
	}
	return name, "component" // fallback
}

// writeIndividualSchemas writes each ConfigMap's JSON schema to the Dagger directory
// and returns categorized schemaEntry values for master schema construction.
func writeIndividualSchemas(dir *dagger.Directory, list SchemaMapList) (*dagger.Directory, []schemaEntry) {
	var entries []schemaEntry

	for _, item := range list.Items {
		schemaContent, ok := item.Data["openapi-v3-json-schema"]
		if !ok || schemaContent == "" {
			continue
		}

		typeName, category := parseEntry(item.Metadata.Name)
		if typeName == "" {
			continue // skip versioned duplicates
		}

		fullName := strings.TrimPrefix(item.Metadata.Name, "schema-")
		fileName := fmt.Sprintf("%s-schema.json", fullName)

		entries = append(entries, schemaEntry{typeName: typeName, fileName: fileName, category: category})
		dir = dir.WithNewFile(fileName, schemaContent)
	}

	return dir, entries
}

// buildConditionals creates allOf if/then rules that activate a $ref schema
// on the 'properties' key when the 'type' field matches a known value.
// Using allOf+if/then instead of oneOf means the base item schema always
// applies and VS Code never gets confused when 'type' hasn't been typed yet.
func buildConditionals(entries []schemaEntry) []map[string]interface{} {
	var rules []map[string]interface{}
	for _, e := range entries {
		rules = append(rules, map[string]interface{}{
			"if": map[string]interface{}{
				"properties": map[string]interface{}{
					"type": map[string]interface{}{"const": e.typeName},
				},
				"required": []string{"type"},
			},
			"then": map[string]interface{}{
				"properties": map[string]interface{}{
					"properties": map[string]interface{}{"$ref": e.fileName},
				},
			},
		})
	}
	return rules
}

// typeEnum extracts an ordered list of type name strings for an enum field.
func typeEnum(entries []schemaEntry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.typeName)
	}
	return names
}

// injectMasterSchema builds a complete, IDE-friendly vela-application-schema.json.
//
// The schema explicitly defines every field a user will encounter in an app.yaml:
//   - metadata.name / namespace / labels / annotations
//   - spec.components[].name, type (enum), properties, traits, dependsOn
//   - spec.components[].traits[].type (enum), properties
//   - spec.policies[].name, type (enum), properties
//   - spec.workflow.steps[].name, type (enum), properties
//
// allOf + if/then rules activate the per-type $ref on 'properties' so that
// after the user writes 'type: webservice', the webservice schema drives hints.
func injectMasterSchema(dir *dagger.Directory, entries []schemaEntry) (*dagger.Directory, error) {
	// ----- bucket entries by category -----
	var compEntries, traitEntries, policyEntries, wfEntries []schemaEntry
	for _, e := range entries {
		switch e.category {
		case "component":
			compEntries = append(compEntries, e)
		case "trait":
			traitEntries = append(traitEntries, e)
		case "policy":
			policyEntries = append(policyEntries, e)
		case "workflowstep":
			wfEntries = append(wfEntries, e)
		}
	}

	// ----- trait item schema -----
	traitItem := map[string]interface{}{
		"type":        "object",
		"description": "An operational capability attached to a component (e.g. ingress, scaler, resource limits).",
		"required":    []string{"type"},
		"properties": map[string]interface{}{
			"type": map[string]interface{}{
				"type":        "string",
				"description": "The trait definition type. Controls which 'properties' are valid.",
				"enum":        typeEnum(traitEntries),
			},
			"properties": map[string]interface{}{
				"type":        "object",
				"description": "Trait-specific properties. Set 'type' first, then use Ctrl+Space here.",
			},
		},
		"allOf": buildConditionals(traitEntries),
	}

	// ----- component item schema -----
	componentItem := map[string]interface{}{
		"type":        "object",
		"description": "A single workload component of the application.",
		"required":    []string{"name", "type"},
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "A unique name for this component instance within the application.",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "The component definition type. Controls which 'properties' are valid. Use Ctrl+Space to see all options.",
				"enum":        typeEnum(compEntries),
			},
			"properties": map[string]interface{}{
				"type":        "object",
				"description": "Component-specific properties. Set 'type' first, then Ctrl+Space here for field hints.",
			},
			"traits": map[string]interface{}{
				"type":        "array",
				"description": "Operational traits to attach to this component (ingress, scaler, resource limits, etc.).",
				"items":       traitItem,
			},
			"dependsOn": map[string]interface{}{
				"type":        "array",
				"description": "Names of other components in the same application that must be deployed first.",
				"items":       map[string]interface{}{"type": "string"},
			},
			"externalRevision": map[string]interface{}{
				"type":        "string",
				"description": "Pin to a specific component revision by name.",
			},
		},
		// allOf applies per-type property schemas without breaking base schema hints
		"allOf": buildConditionals(compEntries),
	}

	// ----- policy item schema -----
	policyItem := map[string]interface{}{
		"type":        "object",
		"description": "An application-level policy (e.g. override, topology, garbage-collect).",
		"required":    []string{"name", "type"},
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "A unique name for this policy.",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "The policy type. Controls which 'properties' are valid.",
				"enum":        typeEnum(policyEntries),
			},
			"properties": map[string]interface{}{
				"type":        "object",
				"description": "Policy-specific properties. Set 'type' first, then Ctrl+Space here.",
			},
		},
		"allOf": buildConditionals(policyEntries),
	}

	// ----- workflow step item schema -----
	workflowStepItem := map[string]interface{}{
		"type":        "object",
		"description": "A single step in the application deployment workflow.",
		"required":    []string{"name", "type"},
		"properties": map[string]interface{}{
			"name": map[string]interface{}{
				"type":        "string",
				"description": "A unique name for this workflow step.",
			},
			"type": map[string]interface{}{
				"type":        "string",
				"description": "The workflow step type. Controls which 'properties' are valid.",
				"enum":        typeEnum(wfEntries),
			},
			"properties": map[string]interface{}{
				"type":        "object",
				"description": "Step-specific properties. Set 'type' first, then Ctrl+Space here.",
			},
			"dependsOn": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "string"},
			},
			"if": map[string]interface{}{
				"type":        "string",
				"description": "A CUE expression. If false, this step is skipped.",
			},
			"timeout": map[string]interface{}{
				"type":        "string",
				"description": "Maximum time to wait for this step, e.g. '10m'.",
			},
			"inputs":  map[string]interface{}{"type": "array"},
			"outputs": map[string]interface{}{"type": "array"},
		},
		"allOf": buildConditionals(wfEntries),
	}

	// ----- root master schema -----
	masterSchema := map[string]interface{}{
		"$schema":     "http://json-schema.org/draft-07/schema#",
		"title":       "KubeVela Application",
		"description": "Defines a KubeVela Application resource (OAM Application). Schemas generated from live cluster.",
		"type":        "object",
		"required":    []string{"apiVersion", "kind", "metadata", "spec"},
		"properties": map[string]interface{}{
			"apiVersion": map[string]interface{}{
				"type":        "string",
				"const":       "core.oam.dev/v1beta1",
				"description": "Must be 'core.oam.dev/v1beta1'.",
			},
			"kind": map[string]interface{}{
				"type":        "string",
				"const":       "Application",
				"description": "Must be 'Application'. Other OAM kinds (ComponentDefinition, TraitDefinition) are for platform operators.",
			},
			"metadata": map[string]interface{}{
				"type":        "object",
				"description": "Standard Kubernetes object metadata.",
				"required":    []string{"name"},
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "The application name. Used as the Kubernetes resource name.",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Kubernetes namespace to deploy into. Defaults to the cluster default.",
					},
					"labels": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": map[string]interface{}{"type": "string"},
						"description":          "Key-value labels attached to this Application resource.",
					},
					"annotations": map[string]interface{}{
						"type":                 "object",
						"additionalProperties": map[string]interface{}{"type": "string"},
						"description":          "Key-value annotations attached to this Application resource.",
					},
				},
			},
			"spec": map[string]interface{}{
				"type":        "object",
				"description": "The desired state of the Application.",
				"required":    []string{"components"},
				"properties": map[string]interface{}{
					"components": map[string]interface{}{
						"type":        "array",
						"description": "The list of workload components that make up this application.",
						"items":       componentItem,
					},
					"policies": map[string]interface{}{
						"type":        "array",
						"description": "Application-level policies such as topology, override, garbage-collect.",
						"items":       policyItem,
					},
					"workflow": map[string]interface{}{
						"type":        "object",
						"description": "Custom deployment workflow. If omitted, KubeVela deploys all components in parallel.",
						"properties": map[string]interface{}{
							"mode": map[string]interface{}{
								"type":        "object",
								"description": "Execution mode for the workflow steps.",
								"properties": map[string]interface{}{
									"steps": map[string]interface{}{
										"type":        "string",
										"enum":        []string{"DAG", "StepByStep"},
										"description": "Top-level step execution order. DAG = parallel where possible, StepByStep = sequential.",
									},
									"subSteps": map[string]interface{}{
										"type":        "string",
										"enum":        []string{"DAG", "StepByStep"},
										"description": "Execution order for steps inside a step-group.",
									},
								},
							},
							"steps": map[string]interface{}{
								"type":        "array",
								"description": "Ordered list of workflow steps.",
								"items":       workflowStepItem,
							},
						},
					},
				},
			},
		},
	}

	masterJSONBytes, err := json.MarshalIndent(masterSchema, "", "  ")
	if err != nil {
		return nil, err
	}

	return dir.WithNewFile("vela-application-schema.json", string(masterJSONBytes)), nil
}
