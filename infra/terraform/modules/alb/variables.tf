variable "project" { type = string }
variable "environment" { type = string }
variable "vpc_id" { type = string }
variable "public_subnet_ids" { type = list(string) }

variable "container_port" {
  type    = number
  default = 8080
}

variable "acm_certificate_arn" { type = string }
