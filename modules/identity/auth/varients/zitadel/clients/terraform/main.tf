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
  # Git-based source for remote backend compatibility
  # Change ref=main to ref=v1.0.0 (or specific tag) for production stability
  source = "git::https://github.com/devsprithvi/nucleus.git//terraform/modules/huggingface/spaces?ref=main"

  name     = "nucleus-zitadel-dev"
  sdk      = "docker"
  private  = false       # Set to true if you want private access
  hardware = "cpu-basic" # Free tier (cpu-upgrade requires paid plan)
  # storage  = "small"   # REMOVED: Requires paid plan, causes 400 error on free tier

  # Environment variables override Zitadel config
  variables = {
    ZITADEL_EXTERNALDOMAIN = "nucleus-zitadel.hf.space"
    ZITADEL_EXTERNALPORT   = "443"
    ZITADEL_EXTERNALSECURE = "true"
  }

  # Secrets (set these via HF Spaces UI or tfvars)
  secrets = {}
}
