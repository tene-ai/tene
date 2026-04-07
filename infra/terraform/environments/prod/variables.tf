variable "project" {
  type    = string
  default = "tene"
}

variable "environment" {
  type    = string
  default = "prod"
}

variable "aws_region" {
  type    = string
  default = "ap-northeast-2"
}

variable "domain_name" {
  type    = string
  default = "tene.sh"
}

variable "vpc_cidr" {
  type    = string
  default = "10.0.0.0/16"
}

variable "availability_zones" {
  type    = list(string)
  default = ["ap-northeast-2a", "ap-northeast-2c"]
}

variable "db_name" {
  type    = string
  default = "tene"
}

variable "db_username" {
  type    = string
  default = "tene_admin"
}

# Sensitive — set via: TF_VAR_db_password or tene run -- terraform apply
variable "db_password" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}

variable "github_org" {
  type    = string
  default = "tomo-kay"
}

variable "github_repo" {
  type    = string
  default = "tene"
}
