package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"dagger/nucleus-ci/internal/dagger"
)

// SchemaMapList represents the expected JSON response from the Kubernetes API
// when listing ConfigMaps that contain KubeVela component/trait/policy schemas.
type SchemaMapList struct {
	Items []SchemaMapItem `json:"items"`
}

// SchemaMapItem represents a single ConfigMap containing one KubeVela schema.
type SchemaMapItem struct {
	Metadata struct {
		Name string `json:"name"`
	} `json:"metadata"`
	Data map[string]string `json:"data"`
}

// ClusterPayload is what executeVPNAndFetch returns — both pieces of cluster data
// in a single round-trip so we only need to connect once.
type ClusterPayload struct {
	// CRDSchema is the raw OpenAPI v3 JSON object from the Application CRD.
	// Stored as json.RawMessage because bash inlines it as a JSON object (not a string).
	CRDSchema json.RawMessage `json:"crd_schema"`

	CompDefSchema         json.RawMessage `json:"comp_def_schema"`
	TraitDefSchema        json.RawMessage `json:"trait_def_schema"`
	PolicyDefSchema       json.RawMessage `json:"policy_def_schema"`
	WorkflowStepDefSchema json.RawMessage `json:"workflow_step_def_schema"`

	// ConfigMaps is the raw JSON list object of per-type property schemas.
	// Stored as json.RawMessage for the same reason.
	ConfigMaps json.RawMessage `json:"config_maps"`
}

// GenerateVelaSchemas fetches both the Application CRD schema (outer structure)
// and the per-type property ConfigMaps (inner detail) from the cluster, then
// merges them into a single vela-application-schema.json for IDE auto-complete.
// Zero fields are written manually — everything comes from the cluster.
func (a *Actions) GenerateVelaSchemas(
	ctx context.Context,
	tailscaleAuthKey *dagger.Secret,
	kubeconfigBase64 *dagger.Secret,
) (*dagger.Directory, error) {
	dag := dagger.Connect()

	// Phase 1: Connect via Tailscale and fetch both data sources in one shot
	rawPayload, err := executeVPNAndFetch(ctx, dag, tailscaleAuthKey, kubeconfigBase64)
	if err != nil {
		return nil, fmt.Errorf("phase 1 (vpn & fetch) failed: %w", err)
	}

	// Phase 2: Decode the JSON envelope
	var payload ClusterPayload
	if err := json.Unmarshal([]byte(rawPayload), &payload); err != nil {
		return nil, fmt.Errorf("phase 2 (decode payload) failed: %w", err)
	}

	// Phase 3: Parse the ConfigMap list for per-type property schemas
	// payload.ConfigMaps is json.RawMessage ([]byte) — pass directly, no conversion needed
	var cmList SchemaMapList
	if err := json.Unmarshal(payload.ConfigMaps, &cmList); err != nil {
		return nil, fmt.Errorf("phase 3 (parse configmaps) failed: %w", err)
	}
	if len(cmList.Items) == 0 {
		return nil, fmt.Errorf("no schema configmaps found in the cluster")
	}

	// Phase 4: Write individual property schema files
	schemaDir := dag.Directory()
	var entries []schemaEntry
	schemaDir, entries = writeIndividualSchemas(schemaDir, cmList)

	// Phase 5: Build master schema using CRD as base, inject per-type property schemas
	schemaDir, err = buildMasterSchema(schemaDir, []byte(payload.CRDSchema), entries)
	if err != nil {
		return nil, fmt.Errorf("phase 5 (build master schema) failed: %w", err)
	}

	// Phase 6: Write the definition CRD schemas and generate the Grandparent schema
	if len(payload.CompDefSchema) > 0 && string(payload.CompDefSchema) != "null" {
		schemaDir = schemaDir.WithNewFile("vela-componentdefinition-schema.json", string(payload.CompDefSchema))
	}
	if len(payload.TraitDefSchema) > 0 && string(payload.TraitDefSchema) != "null" {
		schemaDir = schemaDir.WithNewFile("vela-traitdefinition-schema.json", string(payload.TraitDefSchema))
	}
	if len(payload.PolicyDefSchema) > 0 && string(payload.PolicyDefSchema) != "null" {
		schemaDir = schemaDir.WithNewFile("vela-policydefinition-schema.json", string(payload.PolicyDefSchema))
	}
	if len(payload.WorkflowStepDefSchema) > 0 && string(payload.WorkflowStepDefSchema) != "null" {
		schemaDir = schemaDir.WithNewFile("vela-workflowstepdefinition-schema.json", string(payload.WorkflowStepDefSchema))
	}

	grandparent := `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "anyOf": [
    { "$ref": "vela-application-schema.json" },
    { "$ref": "vela-componentdefinition-schema.json" },
    { "$ref": "vela-traitdefinition-schema.json" },
    { "$ref": "vela-policydefinition-schema.json" },
    { "$ref": "vela-workflowstepdefinition-schema.json" }
  ]
}`
	schemaDir = schemaDir.WithNewFile("vela-grandparent-schema.json", grandparent)

	return schemaDir, nil
}

// executeVPNAndFetch connects via Tailscale and fetches BOTH the Application
// CRD schema and the per-type ConfigMaps in a single cluster session.
// Output is a JSON object: { "crd_schema": "...", "config_maps": "..." }
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

		# ── PHASE 3: Peer status ───────────────────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 3: Checking Tailscale peer status ━━━" >&2
		tailscale status >&2
		echo "" >&2

		# Decode kubeconfig
		echo "$KUBECONFIG_BASE64" | base64 -d > /tmp/kubeconfig
		sed -i 's/certificate-authority-data:.*/insecure-skip-tls-verify: true/g' /tmp/kubeconfig
		export KUBECONFIG=/tmp/kubeconfig

		CLUSTER_IP=$(grep "server:" /tmp/kubeconfig | awk '{print $2}' | head -1 | sed 's|https\?://||' | cut -d: -f1)
		CLUSTER_PORT=$(grep "server:" /tmp/kubeconfig | awk '{print $2}' | head -1 | sed 's|https\?://||' | cut -d: -f2)
		echo "  Target cluster  : $CLUSTER_IP:$CLUSTER_PORT" >&2

		# ── PHASE 4: Tailnet ping ──────────────────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 4: Pinging cluster IP over Tailnet ━━━" >&2
		tailscale ping --timeout=15s "$CLUSTER_IP" >&2 || \
			echo "  ⚠️  ping failed — will still attempt SOCKS5" >&2

		sleep 5

		# ── PHASE 5: SOCKS5 connectivity ──────────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 5: Testing SOCKS5 proxy connectivity ━━━" >&2
		CURL_OUT=$(curl --silent --max-time 10 \
			--proxy socks5://localhost:1055 \
			--insecure \
			--write-out "HTTP_STATUS:%{http_code}" \
			"https://$CLUSTER_IP:$CLUSTER_PORT/healthz" 2>&1 || true)
		echo "  curl response: $CURL_OUT" >&2
		if echo "$CURL_OUT" | grep -qE "HTTP_STATUS:(200|401|403)"; then
			echo "  ✅ SOCKS5 proxy is routing to the cluster API server!" >&2
		else
			die "SOCKS5 proxy connectivity check failed — kubectl will not work"
		fi

		# ── PHASE 6: kubectl check ─────────────────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 6: Routing kubectl through SOCKS5 proxy ━━━" >&2
		export HTTPS_PROXY=socks5://localhost:1055
		export HTTP_PROXY=socks5://localhost:1055
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

		# ── PHASE 7: Fetch ALL data sources ──────────────────────────────────────
		echo "" >&2
		echo "━━━ PHASE 7: Fetching Application and Definition CRDs + ConfigMaps ━━━" >&2

		# Source 1: The Application CRD — provides the FULL outer schema
		# (components[].name, type, traits, dependsOn, policies, workflow — everything)
		# No manual writing needed: the CRD already defines all of this.
		echo "  Fetching Application CRD schema..." >&2
		CRD_JSON=$(kubectl get crd applications.core.oam.dev \
			-o jsonpath='{.spec.versions[?(@.name=="v1beta1")].schema.openAPIV3Schema}' \
			2>/dev/null || \
			kubectl get crd applications.core.oam.dev \
				-o jsonpath='{.spec.versions[0].schema.openAPIV3Schema}' \
			2>/dev/null)

		if [ -z "$CRD_JSON" ]; then
			die "Could not fetch Application CRD schema from cluster"
		fi
		echo "  ✅ Application CRD schema fetched ($(echo "$CRD_JSON" | wc -c) bytes)" >&2

		# Source 2: The Definition CRDs (Component, Trait, Policy, WorkflowStep)
		echo "  Fetching Definition CRD schemas..." >&2
		COMP_DEF_JSON=$(kubectl get crd componentdefinitions.core.oam.dev -o jsonpath='{.spec.versions[?(@.name=="v1beta1")].schema.openAPIV3Schema}' 2>/dev/null || kubectl get crd componentdefinitions.core.oam.dev -o jsonpath='{.spec.versions[0].schema.openAPIV3Schema}' 2>/dev/null || echo "null")
		TRAIT_DEF_JSON=$(kubectl get crd traitdefinitions.core.oam.dev -o jsonpath='{.spec.versions[?(@.name=="v1beta1")].schema.openAPIV3Schema}' 2>/dev/null || kubectl get crd traitdefinitions.core.oam.dev -o jsonpath='{.spec.versions[0].schema.openAPIV3Schema}' 2>/dev/null || echo "null")
		POLICY_DEF_JSON=$(kubectl get crd policydefinitions.core.oam.dev -o jsonpath='{.spec.versions[?(@.name=="v1beta1")].schema.openAPIV3Schema}' 2>/dev/null || kubectl get crd policydefinitions.core.oam.dev -o jsonpath='{.spec.versions[0].schema.openAPIV3Schema}' 2>/dev/null || echo "null")
		WORKFLOW_STEP_DEF_JSON=$(kubectl get crd workflowstepdefinitions.core.oam.dev -o jsonpath='{.spec.versions[?(@.name=="v1beta1")].schema.openAPIV3Schema}' 2>/dev/null || kubectl get crd workflowstepdefinitions.core.oam.dev -o jsonpath='{.spec.versions[0].schema.openAPIV3Schema}' 2>/dev/null || echo "null")
		echo "  ✅ Definition CRD schemas fetched" >&2

		# Source 3: Per-type property detail schemas from ConfigMaps
		echo "  Fetching KubeVela property ConfigMaps..." >&2
		CM_JSON=$(kubectl get configmap -n vela-system -l definition.oam.dev=schema -o json)
		echo "  ✅ ConfigMaps fetched" >&2

		# Emit a single JSON envelope so Go gets everything in one stdout read
		printf '{"crd_schema":%s,"config_maps":%s,"comp_def_schema":%s,"trait_def_schema":%s,"policy_def_schema":%s,"workflow_step_def_schema":%s}' "$CRD_JSON" "$CM_JSON" "${COMP_DEF_JSON:-null}" "${TRAIT_DEF_JSON:-null}" "${POLICY_DEF_JSON:-null}" "${WORKFLOW_STEP_DEF_JSON:-null}"
	`

	return container.
		WithExec([]string{"bash", "-c", script}).
		Stdout(ctx)
}

// buildRunnerContainer abstracts OS-level dependency installation.
func buildRunnerContainer(dag *dagger.Client) *dagger.Container {
	return dag.Container().
		From("ubuntu:22.04").
		WithExec([]string{"apt-get", "update"}).
		WithExec([]string{"apt-get", "install", "-y", "curl", "ca-certificates", "openssh-client"}).
		WithExec([]string{"bash", "-c", "curl -fsSL https://tailscale.com/install.sh | sh"}).
		WithExec([]string{"bash", "-c", "curl -LO https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl && chmod +x kubectl && mv kubectl /usr/local/bin/"})
}

// ── Schema helpers ────────────────────────────────────────────────────────────

// schemaEntry pairs the short KubeVela type name with its disk filename and category.
type schemaEntry struct {
	typeName string // e.g. "webservice"
	fileName string // e.g. "component-schema-webservice-schema.json"
	category string // "component" | "trait" | "policy" | "workflowstep"
}

// categoryInfo maps ConfigMap name prefixes to definition categories.
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
// Returns empty strings for versioned duplicates (e.g. "webservice-v1") which are skipped.
func parseEntry(configMapName string) (typeName, category string) {
	name := strings.TrimPrefix(configMapName, "schema-")
	for _, ci := range categoryInfo {
		if strings.HasPrefix(name, ci.prefix) {
			shortName := strings.TrimPrefix(name, ci.prefix)
			// Skip "-v1", "-v2" suffixes — identical duplicates of the non-versioned entries
			if idx := strings.LastIndex(shortName, "-v"); idx >= 0 {
				suffix := shortName[idx+2:]
				allDigits := len(suffix) > 0
				for _, ch := range suffix {
					if ch < '0' || ch > '9' {
						allDigits = false
						break
					}
				}
				if allDigits {
					return "", ""
				}
			}
			return shortName, ci.category
		}
	}
	return name, "component"
}

// writeIndividualSchemas writes each ConfigMap schema to the Dagger directory
// and returns categorized entries for the master schema merger.
func writeIndividualSchemas(dir *dagger.Directory, list SchemaMapList) (*dagger.Directory, []schemaEntry) {
	var entries []schemaEntry
	for _, item := range list.Items {
		schemaContent, ok := item.Data["openapi-v3-json-schema"]
		if !ok || schemaContent == "" {
			continue
		}
		typeName, category := parseEntry(item.Metadata.Name)
		if typeName == "" {
			continue
		}
		fullName := strings.TrimPrefix(item.Metadata.Name, "schema-")
		fileName := fmt.Sprintf("%s-schema.json", fullName)
		entries = append(entries, schemaEntry{typeName: typeName, fileName: fileName, category: category})
		dir = dir.WithNewFile(fileName, schemaContent)
	}
	return dir, entries
}

// typeEnum returns the sorted list of type names for an enum field.
func typeEnum(entries []schemaEntry) []string {
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		names = append(names, e.typeName)
	}
	return names
}

// buildConditionals creates allOf if/then rules that activate a $ref on
// 'properties' when the 'type' field matches. Using allOf (not oneOf) means
// the base item fields are always visible regardless of whether type is set.
func buildConditionals(entries []schemaEntry) []map[string]interface{} {
	var rules []map[string]interface{}
	for _, e := range entries {
		rules = append(rules, map[string]interface{}{
			"if": map[string]interface{}{
				"required": []string{"type"},
				"properties": map[string]interface{}{
					"type": map[string]interface{}{"const": e.typeName},
				},
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

// buildMasterSchema takes the raw CRD OpenAPI v3 JSON (the outer Application
// structure, fully defined by the cluster) and merges the per-type property
// schemas into it. The result is written as vela-application-schema.json.
//
// Strategy:
//  1. Parse the CRD schema — this already has components[].name, type, traits,
//     dependsOn, policies, workflow etc. No manual writing involved.
//  2. Walk to spec.components.items and inject:
//     a. An enum on the 'type' field listing all known component types
//     b. allOf if/then rules that point 'properties' to the right schema file
//  3. Repeat (2b) for traits[].type → trait schemas
//  4. Do the same for spec.policies.items and spec.workflow.steps.items
func buildMasterSchema(dir *dagger.Directory, crdJSON []byte, entries []schemaEntry) (*dagger.Directory, error) {
	// Parse the CRD schema as a generic map — we don't need typed structs
	var schema map[string]interface{}
	if err := json.Unmarshal(crdJSON, &schema); err != nil {
		return nil, fmt.Errorf("parse CRD schema: %w", err)
	}

	// Bucket entries by category
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

	// Navigate: spec → properties → components → items
	// We use a helper that safely walks the map without panicking on missing keys
	compItems := dig(schema, "properties", "spec", "properties", "components", "items")
	if compItems != nil {
		// Remove the base generic "properties" to prevent VS Code showing duplicates
		if propsMap, ok := dig(compItems, "properties").(map[string]interface{}); ok {
			delete(propsMap, "properties")
		}

		// Inject enum + conditionals onto the component item schema
		if typeObj := dig(compItems, "properties", "type"); typeObj != nil {
			if m, ok := typeObj.(map[string]interface{}); ok {
				m["enum"] = typeEnum(compEntries)
			}
		}
		injectConditionals(compItems, buildConditionals(compEntries))

		// Inject trait type enum + conditionals into traits.items
		traitItems := dig(compItems, "properties", "traits", "items")
		if traitItems != nil {
			if propsMap, ok := dig(traitItems, "properties").(map[string]interface{}); ok {
				delete(propsMap, "properties")
			}
			if typeObj := dig(traitItems, "properties", "type"); typeObj != nil {
				if m, ok := typeObj.(map[string]interface{}); ok {
					m["enum"] = typeEnum(traitEntries)
				}
			}
			injectConditionals(traitItems, buildConditionals(traitEntries))
		}
	}

	// Navigate: spec → properties → policies → items
	policyItems := dig(schema, "properties", "spec", "properties", "policies", "items")
	if policyItems != nil {
		if propsMap, ok := dig(policyItems, "properties").(map[string]interface{}); ok {
			delete(propsMap, "properties")
		}
		if typeObj := dig(policyItems, "properties", "type"); typeObj != nil {
			if m, ok := typeObj.(map[string]interface{}); ok {
				m["enum"] = typeEnum(policyEntries)
			}
		}
		injectConditionals(policyItems, buildConditionals(policyEntries))
	}

	// Navigate: spec → properties → workflow → properties → steps → items
	wfItems := dig(schema, "properties", "spec", "properties", "workflow", "properties", "steps", "items")
	if wfItems != nil {
		if propsMap, ok := dig(wfItems, "properties").(map[string]interface{}); ok {
			delete(propsMap, "properties")
		}
		if typeObj := dig(wfItems, "properties", "type"); typeObj != nil {
			if m, ok := typeObj.(map[string]interface{}); ok {
				m["enum"] = typeEnum(wfEntries)
			}
		}
		injectConditionals(wfItems, buildConditionals(wfEntries))
	}

	// Add the JSON Schema dialect marker so VS Code recognises the format
	schema["$schema"] = "http://json-schema.org/draft-07/schema#"

	out, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return nil, err
	}

	return dir.WithNewFile("vela-application-schema.json", string(out)), nil
}

// dig safely navigates a nested map[string]interface{} by key path.
// Returns nil if any key is missing or the value is not a map.
func dig(m interface{}, keys ...string) interface{} {
	cur := m
	for _, k := range keys {
		mm, ok := cur.(map[string]interface{})
		if !ok {
			return nil
		}
		cur = mm[k]
	}
	return cur
}

// injectConditionals merges allOf rules into a schema object.
// If the object already has allOf entries (from the CRD), we append to them.
func injectConditionals(schemaObj interface{}, rules []map[string]interface{}) {
	if len(rules) == 0 {
		return
	}
	m, ok := schemaObj.(map[string]interface{})
	if !ok {
		return
	}
	existing, _ := m["allOf"].([]interface{})
	merged := make([]interface{}, 0, len(existing)+len(rules))
	merged = append(merged, existing...)
	for _, r := range rules {
		merged = append(merged, r)
	}
	m["allOf"] = merged
}
