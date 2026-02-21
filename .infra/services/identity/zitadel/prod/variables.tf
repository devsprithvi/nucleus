# ==============================================================================
# ZITADEL PROD - VARIABLES
# ==============================================================================

variable "hf_token" {
  type        = string
  description = "Hugging Face API token"
  sensitive   = true
}

variable "hf_username" {
  type        = string
  description = "Hugging Face username or organization"
}

variable "space_name_prefix" {
  type        = string
  description = "Prefix for the HF Space name"
  default     = "nucleus"
}
