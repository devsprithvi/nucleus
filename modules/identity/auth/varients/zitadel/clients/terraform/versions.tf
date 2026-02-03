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

