# ==============================================================================
# ZITADEL - BASE MODULE VARIABLES
# ==============================================================================

# --- Required ---

variable "hf_token" {
  type        = string
  description = "Hugging Face API token (with write permissions)"
  sensitive   = true
}

variable "hf_username" {
  type        = string
  description = "Hugging Face username or organization"
}

variable "environment" {
  type        = string
  description = "Environment name (dev, staging, prod)"
}

# --- Optional ---

variable "space_name_prefix" {
  type        = string
  description = "Prefix for the HF Space name"
  default     = "nucleus"
}

variable "private" {
  type        = bool
  description = "Whether the Space should be private"
  default     = false
}

# --- SMTP Configuration (Optional) ---
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
