output "api_public_ip" {
  description = "IP publica de la API"
  value       = aws_instance.salesflow_api.public_ip
}

output "db_endpoint" {
  description = "Endpoint de la base de datos"
  value       = aws_db_instance.salesflow_db.endpoint
  sensitive   = true
}