# ==============================================================================
# ZITADEL - PROD ENVIRONMENT
# ==============================================================================
# Thin wrapper around _base/ with production-specific configuration.
# ==============================================================================

module "root" {
  source = "../_base"

  hf_token    = var.hf_token
  hf_username = var.hf_username
  environment = "prod"

  # Prod-specific overrides
  space_name_prefix = var.space_name_prefix
  private           = true
}
