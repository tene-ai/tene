variable "project" { type = string }
variable "environment" { type = string }
variable "domain_name" { type = string }

variable "subdomain_prefix" {
  type    = string
  default = "api"
}

variable "san_prefixes" {
  description = "Additional subdomain prefixes for SAN entries"
  type        = list(string)
  default     = ["app"]
}
