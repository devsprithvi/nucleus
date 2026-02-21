# ==============================================================================
# ZITADEL - BASE MODULE OUTPUTS
# ==============================================================================

output "space_id" {
  description = "Hugging Face Space ID"
  value       = module.space.space_id
}

output "space_url" {
  description = "Hugging Face Space page URL"
  value       = module.space.space_url
}

output "embed_url" {
  description = "Direct access URL for the Zitadel instance"
  value       = module.space.embed_url
}
