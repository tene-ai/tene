locals {
  project     = var.project
  environment = var.environment
  region      = var.aws_region
}

# ── VPC ────────────────────────────────────────────
module "vpc" {
  source = "../../modules/vpc"

  project            = local.project
  environment        = local.environment
  vpc_cidr           = var.vpc_cidr
  availability_zones = var.availability_zones
}

# ── NAT (fck-nat, ~$3/mo) ─────────────────────────
module "nat" {
  source = "../../modules/nat"

  project                = local.project
  environment            = local.environment
  aws_region             = local.region
  vpc_id                 = module.vpc.vpc_id
  vpc_cidr               = var.vpc_cidr
  public_subnet_id       = module.vpc.public_subnet_ids[0]
  private_route_table_id = module.vpc.private_route_table_id
}

# ── ECR ────────────────────────────────────────────
module "ecr" {
  source  = "../../modules/ecr"
  project = local.project
}

# ── ACM + DNS Validation ──────────────────────────
module "acm" {
  source = "../../modules/acm"

  project     = local.project
  environment = local.environment
  domain_name = var.domain_name
}

# ── S3 (vault blobs + ALB logs) ───────────────────
module "s3" {
  source = "../../modules/s3"

  project     = local.project
  environment = local.environment
  aws_region  = local.region
}

# ── Secrets Manager ───────────────────────────────
module "secrets" {
  source = "../../modules/secrets"

  project      = local.project
  environment  = local.environment
  db_password  = var.db_password
  jwt_secret   = var.jwt_secret
}

# ── IAM ───────────────────────────────────────────
module "iam" {
  source = "../../modules/iam"

  project          = local.project
  environment      = local.environment
  vault_bucket_arn = module.s3.vault_bucket_arn
  secrets_arn      = module.secrets.secret_arn
  github_org       = var.github_org
  github_repo      = var.github_repo
}

# ── ALB ───────────────────────────────────────────
module "alb" {
  source = "../../modules/alb"

  project             = local.project
  environment         = local.environment
  vpc_id              = module.vpc.vpc_id
  public_subnet_ids   = module.vpc.public_subnet_ids
  acm_certificate_arn = module.acm.certificate_arn
}

# ── ECS Fargate ───────────────────────────────────
module "ecs" {
  source = "../../modules/ecs"

  project               = local.project
  environment           = local.environment
  aws_region            = local.region
  vpc_id                = module.vpc.vpc_id
  private_subnet_ids    = module.vpc.private_subnet_ids
  ecr_repository_url    = module.ecr.repository_url
  execution_role_arn    = module.iam.ecs_execution_role_arn
  task_role_arn         = module.iam.ecs_task_role_arn
  target_group_arn      = module.alb.target_group_arn
  alb_security_group_id = module.alb.alb_security_group_id
  db_host               = module.rds.address
  db_name               = var.db_name
  s3_bucket_name        = module.s3.vault_bucket_name
  secrets_arn           = module.secrets.secret_arn
}

# ── RDS PostgreSQL ────────────────────────────────
module "rds" {
  source = "../../modules/rds"

  project               = local.project
  environment           = local.environment
  vpc_id                = module.vpc.vpc_id
  isolated_subnet_ids   = module.vpc.isolated_subnet_ids
  ecs_security_group_id = module.ecs.ecs_security_group_id
  db_name               = var.db_name
  db_username           = var.db_username
  db_password           = var.db_password
}

# ── CloudWatch Alarms ─────────────────────────────
module "cloudwatch" {
  source = "../../modules/cloudwatch"

  project     = local.project
  environment = local.environment
  aws_region  = local.region

  ecs_cluster_name        = module.ecs.cluster_name
  ecs_service_name        = module.ecs.service_name
  rds_instance_id         = module.rds.instance_id
  alb_arn_suffix          = module.alb.alb_arn_suffix
  target_group_arn_suffix = module.alb.target_group_arn_suffix
}

# ── Route 53 (api.tene.sh → ALB) ─────────────────
module "route53" {
  source = "../../modules/route53"

  domain_name     = var.domain_name
  route53_zone_id = module.acm.route53_zone_id
  alb_dns_name    = module.alb.alb_dns_name
  alb_zone_id     = module.alb.alb_zone_id
}
