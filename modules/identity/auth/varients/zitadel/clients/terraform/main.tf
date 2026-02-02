# =============================================================================
# Deploy Zitadel (All-in-One) to Hugging Face Spaces
# =============================================================================
# This deploys the all-in-one Dockerfile that combines:
# - PostgreSQL database
# - Zitadel auth server
# - All in a single container using supervisord
#
# NOTE: This is a TEMPORARY/DEVELOPMENT solution. For production, use
# proper separated services or Zitadel Cloud.
# =============================================================================

module "zitadel" {
  source = "../../../../../../../terraform/modules/huggingface/spaces"

  name     = "nucleus-zitadel"
  sdk      = "docker"
  private  = false         # Set to true if you want private access
  hardware = "cpu-upgrade" # Need more resources for PostgreSQL + Zitadel
  storage  = "small"       # Persistent storage for database

  # Environment variables override Zitadel config
  variables = {
    ZITADEL_EXTERNALDOMAIN = "nucleus-zitadel.hf.space"
    ZITADEL_EXTERNALPORT   = "443"
    ZITADEL_EXTERNALSECURE = "true"
  }

  # Secrets (set these via HF Spaces UI or tfvars)
  secrets = {}
}
