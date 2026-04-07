variable "project" { type = string }
variable "environment" { type = string }
variable "aws_region" { type = string }
variable "vpc_id" { type = string }
variable "private_subnet_ids" { type = list(string) }
variable "ecr_repository_url" { type = string }

variable "image_tag" {
  type    = string
  default = "latest"
}

variable "container_port" {
  type    = number
  default = 8080
}

variable "task_cpu" {
  type    = number
  default = 256
}

variable "task_memory" {
  type    = number
  default = 512
}

variable "desired_count" {
  type    = number
  default = 1
}

variable "min_capacity" {
  type    = number
  default = 1
}

variable "max_capacity" {
  type    = number
  default = 3
}

variable "execution_role_arn" { type = string }
variable "task_role_arn" { type = string }
variable "target_group_arn" { type = string }
variable "alb_security_group_id" { type = string }
variable "db_host" { type = string }
variable "db_name" { type = string }
variable "s3_bucket_name" { type = string }
variable "secrets_arn" { type = string }

variable "lemon_store_id" {
  type    = string
  default = ""
}

variable "lemon_variant_pro" {
  type    = string
  default = ""
}
