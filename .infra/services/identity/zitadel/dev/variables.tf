# ==============================================================================
# ZITADEL DEV - VARIABLES
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

# --- Optional SMTP ---
variable "smtp_host" {
  type        = string
  description = "SMTP Host (e.g. smtp.gmail.com:587)"
  default     = ""
}

variable "smtp_user" {
  type        = string
  description = "SMTP Username"
  default     = ""
}

variable "smtp_password" {
  type        = string
  description = "SMTP Password"
  sensitive   = true
  default     = ""
}
