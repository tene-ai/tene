variable "project" { type = string }
variable "environment" { type = string }

variable "db_password" {
  type      = string
  sensitive = true
}

variable "jwt_secret" {
  type      = string
  sensitive = true
}

variable "lemon_api_key" {
  type      = string
  sensitive = true
  default   = ""
}

variable "lemon_webhook_secret" {
  type      = string
  sensitive = true
  default   = ""
}

variable "github_client_id" {
  type      = string
  sensitive = true
  default   = ""
}

variable "github_client_secret" {
  type      = string
  sensitive = true
  default   = ""
}
