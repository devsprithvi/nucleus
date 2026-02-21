# ==============================================================================
# ZITADEL - BASE MODULE (Root Module)
# ==============================================================================
# This is the single source of truth for Zitadel's infrastructure.
# Environments (dev/, prod/) instantiate this module with their own variables.
#
# Deployment Target: Hugging Face Spaces (Docker SDK, Free Tier)
# Strategy:          local-exec (no official HF Terraform provider)
# ==============================================================================

# ------------------------------------------------------------------------------
# Hugging Face Space
# ------------------------------------------------------------------------------

locals {
  space_name = "${var.space_name_prefix}-zitadel-${var.environment}"
}

module "space" {
  source = "../../../../../terraform/modules/huggingface/spaces"

  hf_token    = var.hf_token
  hf_username = var.hf_username
  name        = local.space_name
  sdk         = "docker"
  private     = var.private

  # Push the Zitadel all-in-one Dockerfile to the Space
  dockerfile_path = "${path.module}/../../../../../services/identity/auth/varients/zitadel/client/.docker/Dockerfile"

  variables = merge({
    # Tell Zitadel its real external domain (HF Spaces proxies HTTPS → port 7860)
    ZITADEL_EXTERNALDOMAIN = "${var.hf_username}-${local.space_name}.hf.space"
    ZITADEL_EXTERNALSECURE = "true"
    ZITADEL_EXTERNALPORT   = "443"

    # Optional SMTP configuration
    }, var.smtp_host != "" ? {
    ZITADEL_NOTIFICATIONS_EMAIL_SMTP_HOST     = var.smtp_host
    ZITADEL_NOTIFICATIONS_EMAIL_SMTP_USER     = var.smtp_user
    ZITADEL_NOTIFICATIONS_EMAIL_SMTP_PASSWORD = var.smtp_password
    ZITADEL_NOTIFICATIONS_EMAIL_SMTP_TLS      = "true"
    ZITADEL_NOTIFICATIONS_EMAIL_SMTP_SENDER   = var.smtp_user
  } : {})
}

