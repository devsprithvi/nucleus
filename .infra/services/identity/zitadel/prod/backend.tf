# ==============================================================================
# ZITADEL PROD - BACKEND CONFIGURATION
# ==============================================================================

terraform {
  backend "remote" {
    workspaces {
      name = "prod-identity-zitadel"
    }
  }

  required_version = ">= 1.0.0"
}
