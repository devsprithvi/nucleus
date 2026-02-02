# ==============================================================================
# REQUIRED
# ==============================================================================

variable "name" {
  type        = string
  description = "Name of the Space"
}

# ==============================================================================
# OPTIONAL
# ==============================================================================

variable "private" {
  type        = bool
  description = "Whether the Space is private"
  default     = false
}

variable "sdk" {
  type        = string
  description = "SDK type: docker, gradio, streamlit, static"
  default     = "docker"
}

variable "template" {
  type        = string
  description = "Template to use (e.g., zenml/zenml)"
  default     = null
}

variable "hardware" {
  type        = string
  description = "Hardware tier: cpu-basic, cpu-upgrade, t4-small, t4-medium, a10g-small, a10g-large, a100-large"
  default     = "cpu-basic"
}

variable "storage" {
  type        = string
  description = "Storage size: small, medium, large"
  default     = null
}

variable "sleep_time" {
  type        = number
  description = "Seconds before sleep (0 = never)"
  default     = 3600
}

variable "secrets" {
  type        = map(string)
  description = "Secret environment variables"
  default     = {}
  sensitive   = true
}

variable "variables" {
  type        = map(string)
  description = "Environment variables"
  default     = {}
}
