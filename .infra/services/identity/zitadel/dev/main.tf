# ==============================================================================
# ZITADEL - DEV ENVIRONMENT
# ==============================================================================
# Thin wrapper around _base/ with dev-specific configuration.
#
# Usage:
#   cd .infra/services/identity/zitadel/dev/
#   tofu init -backend-config=../../../../.shared/config/backend.hcl
#   tofu plan -var-file=terraform.tfvars
#   tofu apply -var-file=terraform.tfvars
# ==============================================================================

module "root" {
  source = "../_base"

  hf_token    = var.hf_token
  hf_username = var.hf_username
  environment = "dev"

  # Dev-specific overrides
  space_name_prefix = var.space_name_prefix
  private           = false

  # Optional SMTP
  smtp_host     = var.smtp_host
  smtp_user     = var.smtp_user
  smtp_password = var.smtp_password
}
