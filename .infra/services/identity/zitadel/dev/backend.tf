# ==============================================================================
# ZITADEL DEV - BACKEND CONFIGURATION
# ==============================================================================
# NOTE: Using local backend for initial development/testing.
# Migrate to remote (Scalr) once workspace variables are configured:
#
#   terraform {
#     backend "remote" {
#       workspaces { name = "dev-identity-zitadel" }
#     }
#   }
# ==============================================================================

terraform {
  backend "local" {
    path = "terraform.tfstate"
  }

  required_version = ">= 1.0.0"
}
