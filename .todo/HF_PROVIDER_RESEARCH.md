# Engineering Log: Hugging Face Terraform Provider Strategy

**Date:** 2023-10-27
**Status:** BLOCKED (Requires Custom Development)
**Author:** [Your Name]

## 1. The Core Problem
We need to manage Hugging Face Spaces (Docker/Gradio) as Infrastructure-as-Code.
- **Current State:** Using `local-exec` scripts (fragile, no state management).
- **Desired State:** Native Terraform Resource `resource "huggingface_space" "main" {...}`.

## 2. Market Research (Dead Ends)
*I have already investigated the following paths. DO NOT RE-INVESTIGATE.*

### A. Community Providers
- **Repo:** `github.com/some-user/terraform-provider-huggingface`
- **Verdict:** **UNUSABLE.**
- **Reason:** Last commit was 2 years ago. Lacks support for "Spaces" (only supports Model Hub). API endpoints are 404.

### B. OpenAPI Generators
- **Hypothesis:** Generate a provider automatically from HF OpenAPI specs.
- **Experiment:** Ran `openapi-generator-cli` against `huggingface.co/api/openapi.json`.
- **Result:** **FAILED.**
    - The HF OpenAPI spec is incomplete regarding authentication scopes for Spaces.
    - The generated Go code requires 40% manual rewrite.
    - **Conclusion:** Automated generation is a rabbit hole. Abandon this path.

## 3. The Technical Solution (The Bridge)
We must build a **Custom Provider** using the HashiCorp Plugin Framework.

### Architecture Plan
1.  **SDK:** Use the official Go SDK `github.com/huggingface/hf-hub` (It has good coverage).
2.  **Resource Schema:**
    ```go
    schema.Schema{
        "name": schema.StringAttribute{Required: true},
        "sdk":  schema.StringAttribute{Optional: true, Default: "docker"},
        "hardware": schema.StringAttribute{Optional: true, Default: "cpu-basic"},
    }
    ```
3.  **Authentication:** Must handle HF_TOKEN env var directly.

## 4. Next Actions (The Chain)
I cannot start coding this provider without a proper project tracker because this is a multi-week effort.
**BLOCKER:** We need a Jira project configured to track "Custom Provider Dev".
**SEE:** `.todo/JIRA_SETUP_SPECS.md`