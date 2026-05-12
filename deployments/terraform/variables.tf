variable "aws_region" {
  description = "Region de AWS"
  default     = "us-east-1"
}

variable "instance_type" {
  description = "Tipo de instancia EC2"
  default     = "t2.micro"
}

variable "db_name" {
  description = "Nombre de la base de datos"
  default     = "salesflow_db"
}

variable "db_user" {
  description = "Usuario de la base de datos"
  default     = "salesflow"
}

variable "db_password" {
  description = "Password de la base de datos"
  sensitive   = true
}