terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.40"
    }
  }

  # To activate: run bootstrap first (see infra/terraform/bootstrap/)
  # then uncomment this block and run: terraform init -migrate-state
  backend "s3" {
    bucket         = "tene-terraform-state-ap-northeast-2"
    key            = "prod/terraform.tfstate"
    region         = "ap-northeast-2"
    dynamodb_table = "tene-terraform-lock"
    encrypt        = true
    profile        = "monsa-sandbox"
  }
}

provider "aws" {
  region  = var.aws_region
  profile = "monsa-sandbox"

  default_tags {
    tags = {
      Project     = var.project
      Environment = var.environment
      ManagedBy   = "terraform"
    }
  }
}
