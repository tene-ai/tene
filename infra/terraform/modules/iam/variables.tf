variable "project" { type = string }
variable "environment" { type = string }
variable "vault_bucket_arn" { type = string }
variable "secrets_arn" { type = string }
variable "github_org" { type = string }
variable "github_repo" { type = string }

variable "create_oidc_provider" {
  description = "Whether to create the GitHub OIDC provider (only needed once per account)"
  type        = bool
  default     = true
}
