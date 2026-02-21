# ==============================================================================
# ZITADEL DEV - OUTPUTS
# ==============================================================================

output "space_id" {
  description = "Hugging Face Space ID"
  value       = module.root.space_id
}

output "space_url" {
  description = "Hugging Face Space page URL"
  value       = module.root.space_url
}

output "embed_url" {
  description = "Direct access URL for the Zitadel dev instance"
  value       = module.root.embed_url
}
