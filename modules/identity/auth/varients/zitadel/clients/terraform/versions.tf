terraform {
  required_version = ">= 1.0.0"

  # =============================================================================
  # REMOTE BACKEND (Scalr)
  # =============================================================================
  # IMPORTANT: The remote backend will NOT work with relative local module paths!
  # 
  # Before enabling this backend, you MUST:
  # 1. Push this repo to Git (GitHub, GitLab, etc.)
  # 2. Change the module source in main.tf from relative path to git-based:
  #    
  #    source = "git::https://github.com/YOUR_ORG/nucleus.git//terraform/modules/huggingface/spaces?ref=main"
  #
  # Then uncomment the backend block below.
  # =============================================================================
  # backend "remote" {
  #   hostname     = "devsprithvi.scalr.io"
  #   organization = "env-v0p23p6v5h5ug56ti"
  #   workspaces {
  #     name = "zitadel-clients"
  #   }
  # }

  # Using Terraform registry as OpenTofu registry doesn't have this provider
  required_providers {
    huggingface-spaces = {
      source  = "registry.terraform.io/strickvl/huggingface-spaces"
      version = "0.0.4"
    }
  }
}

