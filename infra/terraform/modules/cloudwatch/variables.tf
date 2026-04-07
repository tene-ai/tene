variable "project" { type = string }
variable "environment" { type = string }
variable "aws_region" { type = string }

# ECS
variable "ecs_cluster_name" { type = string }
variable "ecs_service_name" { type = string }

# RDS
variable "rds_instance_id" {
  type    = string
  default = ""
}

# ALB
variable "alb_arn_suffix" {
  type    = string
  default = ""
}

variable "target_group_arn_suffix" {
  type    = string
  default = ""
}

# SNS (optional — set to "" to skip notifications)
variable "sns_topic_arn" {
  type    = string
  default = ""
}
