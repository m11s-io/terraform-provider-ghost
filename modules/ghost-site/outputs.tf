output "integration_id" {
  description = "Ghost-assigned ID of the managed integration."
  value       = ghost_integration.this.id
}

output "content_api_key" {
  description = "Content API key for this Ghost instance."
  value       = ghost_integration.this.content_api_key
  sensitive   = true
}

output "admin_api_key" {
  description = "Admin API key for this Ghost instance (id:hex format)."
  value       = ghost_integration.this.admin_api_key
  sensitive   = true
}
