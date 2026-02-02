# ==============================================================================
# OUTPUTS
# Note: The huggingface_endpoint resource does not export id, url, or status
# attributes. We output the input values for reference.
# ==============================================================================

output "name" {
  description = "Endpoint name"
  value       = huggingface_endpoint.this.name
}

output "type" {
  description = "Endpoint type (public/private)"
  value       = huggingface_endpoint.this.type
}

output "cloud_vendor" {
  description = "Cloud vendor"
  value       = var.cloud_vendor
}

output "cloud_region" {
  description = "Cloud region"
  value       = var.cloud_region
}

output "model_repository" {
  description = "Model repository"
  value       = var.model_repository
}
