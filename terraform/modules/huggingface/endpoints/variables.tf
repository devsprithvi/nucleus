# ==============================================================================
# REQUIRED
# ==============================================================================

variable "name" {
  type        = string
  description = "Name of the endpoint"
}

variable "model_repository" {
  type        = string
  description = "Hugging Face model repository (e.g., meta-llama/Llama-2-7b)"
}

# ==============================================================================
# OPTIONAL - GENERAL
# ==============================================================================

variable "type" {
  type        = string
  description = "Endpoint type: public, private"
  default     = "public"
}

# ==============================================================================
# OPTIONAL - CLOUD
# ==============================================================================

variable "cloud_vendor" {
  type        = string
  description = "Cloud provider"
  default     = "aws"
}

variable "cloud_region" {
  type        = string
  description = "Cloud region"
  default     = "us-east-1"
}

# ==============================================================================
# OPTIONAL - COMPUTE
# ==============================================================================

variable "accelerator" {
  type        = string
  description = "Accelerator type: cpu, gpu"
  default     = "cpu"
}

variable "instance_type" {
  type        = string
  description = "Instance type (e.g., c6i, nvidia-l4, nvidia-a100)"
  default     = "c6i"
}

variable "instance_size" {
  type        = string
  description = "Instance size: x1, x2, x4, x8"
  default     = "x1"
}

variable "min_replicas" {
  type        = number
  description = "Minimum replicas (0 for scale-to-zero)"
  default     = 0
}

variable "max_replicas" {
  type        = number
  description = "Maximum replicas"
  default     = 1
}

variable "scale_to_zero_timeout" {
  type        = number
  description = "Seconds before scaling to zero"
  default     = 900
}

# ==============================================================================
# OPTIONAL - MODEL
# ==============================================================================

variable "model_task" {
  type        = string
  description = "Model task: text-generation, text-classification, etc."
  default     = "text-generation"
}

variable "model_framework" {
  type        = string
  description = "Framework: pytorch, tensorflow, onnx"
  default     = "pytorch"
}

variable "model_revision" {
  type        = string
  description = "Model revision/branch"
  default     = "main"
}

# ==============================================================================
# OPTIONAL - IMAGE
# ==============================================================================

variable "image_huggingface" {
  type        = object({})
  description = "Hugging Face native image configuration (use empty object {} for default)"
  default     = {}
}
