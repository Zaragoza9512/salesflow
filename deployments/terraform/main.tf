terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
  backend "s3" {
    bucket = "salesflow-terraform-state"
    key    = "prod/terraform.tfstate"
    region = "us-east-1"
  }
}

provider "aws" {
  region = var.aws_region
}

# EC2 para la API
resource "aws_instance" "salesflow_api" {
  ami           = "ami-0c02fb55956c7d316"
  instance_type = var.instance_type

  tags = {
    Name = "salesflow-api"
  }
}

# RDS PostgreSQL
resource "aws_db_instance" "salesflow_db" {
  identifier        = "salesflow-db"
  engine            = "postgres"
  engine_version    = "16"
  instance_class    = "db.t3.micro"
  allocated_storage = 20
  db_name           = var.db_name
  username          = var.db_user
  password          = var.db_password
  skip_final_snapshot = true

  tags = {
    Name = "salesflow-db"
  }
}