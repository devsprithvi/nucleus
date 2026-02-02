output "id" {
  description = "Space ID"
  value       = huggingface-spaces_space.this.id
}

output "name" {
  description = "Space name"
  value       = huggingface-spaces_space.this.name
}

output "url" {
  description = "Space URL"
  value       = "https://huggingface.co/spaces/${huggingface-spaces_space.this.id}"
}
