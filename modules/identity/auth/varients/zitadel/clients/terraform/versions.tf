terraform {
  required_version = ">= 1.0.0"

  # =============================================================================
  # REMOTE BACKEND (Scalr)
  # =============================================================================
  # State is managed remotely via Scalr. Module sources use Git-based paths.
  # 
  # Version Control:
  #   - ref=main       → Development (latest, may change)
  #   - ref=v1.0.0     → Production (stable, versioned)
  # =============================================================================
  backend "remote" {
    hostname     = "devsprithvi.scalr.io"
    organization = "env-v0p23p6v5h5ug56ti"
    workspaces {
      name = "zitadel-clients"
    }
  }

  # Using Terraform registry as OpenTofu registry doesn't have this provider
  required_providers {
    huggingface-spaces = {
      source  = "registry.terraform.io/strickvl/huggingface-spaces"
      version = "0.0.4"
    }
  }
}




# =============================================================================
# PROVIDER CONFIGURATION (Managed by Scalr)
# =============================================================================
# The huggingface-spaces provider is configured via Scalr's Provider 
# Configuration feature. Scalr automatically injects the provider block.
#
# Setup in Scalr:
#   1. Go to Scalr > Account > Provider Configurations
#   2. Click "+ New Provider Configuration"
#   3. Select "Custom" (since huggingface-spaces is not a built-in provider)
#   4. Configure:
#      - Provider Name: huggingface-spaces
#      - Add argument: token = <your-huggingface-write-token>
#   5. Link to workspace or set as environment default
#
# To get a HuggingFace token:
#   - Go to https://huggingface.co/settings/tokens
#   - Create a new token with "Write" access
# =============================================================================
