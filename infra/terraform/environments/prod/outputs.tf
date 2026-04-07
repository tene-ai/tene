output "vpc_id" { value = module.vpc.vpc_id }
output "ecr_repository_url" { value = module.ecr.repository_url }
output "alb_dns_name" { value = module.alb.alb_dns_name }
output "rds_endpoint" {
  value     = module.rds.endpoint
  sensitive = true
}
output "ecs_cluster" { value = module.ecs.cluster_name }
output "ecs_service" { value = module.ecs.service_name }
output "github_actions_role_arn" { value = module.iam.github_actions_role_arn }
output "s3_vault_bucket" { value = module.s3.vault_bucket_name }
