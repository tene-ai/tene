# Tene AWS Infrastructure Report

**Date**: 2026-04-07
**Account**: 507221376909 (monsa-sandbox)
**Region**: ap-northeast-2 (Seoul)
**Prepared by**: Kay

---

## 1. Compute

| Resource | Prod | Staging | Pricing Basis |
|----------|------|---------|---------------|
| **ECS Fargate** | 0.25 vCPU / 512 MiB x **2 tasks** | 0.25 vCPU / 512 MiB x **1 task** | vCPU $0.04048/h, Memory $0.00442/h/GB |
| ECS Auto Scaling | min 2 / max 3 | min 1 / max 2 | |
| **fck-nat (EC2)** | t4g.nano x 1 | t4g.nano x 1 | $0.0042/h |
| **Elastic IP** | 1 (3.39.164.212) | 1 (43.202.235.193) | Associated = free |

## 2. Database

| Resource | Prod | Staging |
|----------|------|---------|
| **RDS PostgreSQL 16.6** | db.t4g.micro | db.t4g.micro |
| vCPU / RAM | 2 vCPU / 1 GB | 2 vCPU / 1 GB |
| Multi-AZ | **Enabling** (set to true) | false |
| Storage | 20 GB gp3 (max 100 auto-scale) | 20 GB gp3 (max 100 auto-scale) |
| Encrypted | AES-256 | AES-256 |
| Backup Retention | 7 days | 7 days |
| Performance Insights | Enabled | Enabled |
| Deletion Protection | Enabled | Enabled |

## 3. Networking

| Resource | Prod | Staging |
|----------|------|---------|
| **VPC** | 10.0.0.0/16 | 10.1.0.0/16 |
| Public Subnets | 2 (2a, 2c) | 2 (2a, 2c) |
| Private Subnets | 2 (2a, 2c) | 2 (2a, 2c) |
| Isolated Subnets | 2 (2a, 2c) | 2 (2a, 2c) |
| Internet Gateway | 1 | 1 |
| Security Groups | 4 (ALB/ECS/RDS/NAT) | 4 (ALB/ECS/RDS/NAT) |
| **ALB** | internet-facing, 2 AZ | internet-facing, 2 AZ |
| ALB Target Group | ip type, port 8080 | ip type, port 8080 |
| ALB Deletion Protection | Enabled | Disabled |

## 4. DNS & TLS

| Resource | Detail |
|----------|--------|
| **Route53 Hosted Zone** | tene.sh (18 records, shared) |
| **ACM Certificate (prod)** | api.tene.sh -- ISSUED |
| **ACM Certificate (staging)** | api-staging.tene.sh -- ISSUED |

## 5. Storage

| Bucket | Env | Usage |
|--------|-----|-------|
| tene-vault-prod-ap-northeast-2 | prod | 0 objects |
| tene-vault-staging-ap-northeast-2 | staging | 0 objects |
| tene-alb-logs-prod-ap-northeast-2 | prod | 0 objects |
| tene-alb-logs-staging-ap-northeast-2 | staging | 0 objects |
| tene-terraform-state-ap-northeast-2 | shared | 310 KB (2 state files) |

## 6. Security & IAM

| Resource | Prod | Staging |
|----------|------|---------|
| **Secrets Manager** | tene/prod/api-secrets (6 keys) | tene/staging/api-secrets (6 keys) |
| IAM Role -- ECS Execution | tene-prod-ecs-execution | tene-staging-ecs-execution |
| IAM Role -- ECS Task | tene-prod-ecs-task | tene-staging-ecs-task |
| IAM Role -- GitHub Actions | tene-prod-github-actions | tene-staging-github-actions |
| OIDC Provider | token.actions.githubusercontent.com (shared) | |
| **ECR** | tene-api (IMMUTABLE tags, scan on push, shared) -- 10 images | |

All resources tagged with: `Project=tene`, `Environment=prod/staging`, `ManagedBy=terraform`

## 7. Observability

| Resource | Prod | Staging |
|----------|------|---------|
| CloudWatch Log Group | /ecs/tene-prod/api (30d retention) | /ecs/tene-staging/api (30d retention) |
| CloudWatch Alarms | **11** (all OK) | **11** (all OK) |
| DynamoDB (TF Lock) | tene-terraform-lock (PAY_PER_REQUEST, shared) | |

### CloudWatch Alarms (per environment)

- ECS: CPU high, Memory high, No running tasks
- RDS: CPU high, Low storage, Connections high
- ALB: 5xx errors, Target 5xx, High latency (P99), Unhealthy hosts
- NAT: Auto recovery

---

## Monthly Cost Estimate (ap-northeast-2)

| Service | Prod | Staging | Calculation |
|---------|-----:|--------:|-------------|
| **ALB** | $22.58 | $22.58 | $0.0288/h x 730h + LCU minimum |
| **ECS Fargate** | $14.90 | $7.45 | (0.25vCPU x $0.04048 + 0.5GB x $0.00442) x 730h x tasks |
| **RDS (Multi-AZ off)** | $11.68 | $11.68 | db.t4g.micro $0.016/h x 730h |
| **RDS (Multi-AZ on)** | $23.36 | -- | x2 when Multi-AZ completes |
| **RDS Storage** | $2.40 | $2.40 | 20GB x $0.12/GB/mo (gp3) |
| **fck-nat (EC2)** | $3.07 | $3.07 | t4g.nano $0.0042/h x 730h |
| **EIP** | $0 | $0 | Associated = free |
| **S3** | ~$0.05 | ~$0.05 | Minimal usage |
| **Secrets Manager** | $0.40 | $0.40 | $0.40/secret/mo |
| **Route53** | $0.50 | -- | 1 hosted zone (shared) |
| **ECR** | ~$0.10 | -- | 10 images, shared |
| **CloudWatch Logs** | ~$0.10 | ~$0.10 | Minimal usage |
| **CloudWatch Alarms** | $1.10 | $1.10 | $0.10/alarm x 11 |
| **DynamoDB** | ~$0 | -- | PAY_PER_REQUEST, shared |
| **ACM** | $0 | $0 | Free |
| | | | |
| **Subtotal (Multi-AZ off)** | **$56.88** | **$48.83** | |
| **Subtotal (Multi-AZ on)** | **$68.56** | -- | |

### Total

| Scenario | Monthly Cost |
|----------|-------------:|
| **Current (Multi-AZ off)** | **$105.71/mo** |
| **After Multi-AZ enabled** | **$117.39/mo** |

> **Note**: Prod RDS Multi-AZ has been configured in Terraform (set to true) but AWS is still applying the change. Once complete, the prod RDS cost doubles from $11.68 to $23.36 (+$11.68/mo).

---

## Architecture Diagram

```
                    Internet
                       |
                  Route 53 (tene.sh)
                   /         \
           api.tene.sh    api-staging.tene.sh
               |                  |
          [Prod ALB]        [Staging ALB]
          (2 AZ, HTTPS)     (2 AZ, HTTPS)
               |                  |
        +-----------+        +--------+
        | ECS x2    |        | ECS x1 |
        | (Fargate) |        |(Fargate)|
        +-----------+        +--------+
               |                  |
          [Prod RDS]        [Staging RDS]
          (Multi-AZ)        (Single-AZ)
          db.t4g.micro      db.t4g.micro

Shared: ECR (tene-api), OIDC Provider, Route53 Zone, TF State (S3+DynamoDB)
```

## Endpoints

| Service | Prod | Staging |
|---------|------|---------|
| API | https://api.tene.sh | https://api-staging.tene.sh |
| Dashboard | https://app.tene.sh (Vercel) | https://app-staging.tene.sh (Vercel) |
| Landing | https://tene.sh (Vercel) | -- |

## Terraform Structure

```
infra/terraform/
  environments/
    prod/       -- S3 backend (prod/terraform.tfstate)
    staging/    -- S3 backend (staging/terraform.tfstate)
  modules/      -- 12 shared modules
    vpc, nat, ecs, alb, rds, s3, ecr, acm, route53, iam, secrets, cloudwatch
```

## Git Branch Strategy

| Branch | Deploy Target | Protection |
|--------|---------------|------------|
| main | Prod (auto-deploy via CI) | Force push blocked, deletion blocked |
| staging | Staging (auto-deploy via CI) | Force push blocked, deletion blocked |
| feature/* | PR -> Vercel Preview + CI | -- |
