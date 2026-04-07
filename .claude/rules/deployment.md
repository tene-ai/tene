# Tene Deployment Guide

## Deployment Checklist

### Before Deploying

1. `go test ./...` — 전체 테스트 통과 확인
2. `go build ./...` — 빌드 성공 확인
3. `tene list --env prod` — prod 시크릿 최신 상태 확인

### API Server (Go → ECS)

```bash
# 1. Docker 이미지 빌드
docker build --platform linux/amd64 -t tene-api -f Dockerfile.server .

# 2. ECR 로그인 + 푸시
aws ecr get-login-password --region ap-northeast-2 --profile monsa-sandbox | \
  docker login --username AWS --password-stdin 507221376909.dkr.ecr.ap-northeast-2.amazonaws.com

IMAGE_TAG="v$(date +%Y%m%d-%H%M%S)"
docker tag tene-api:latest 507221376909.dkr.ecr.ap-northeast-2.amazonaws.com/tene-api:$IMAGE_TAG
docker push 507221376909.dkr.ecr.ap-northeast-2.amazonaws.com/tene-api:$IMAGE_TAG

# 3. ECS 배포
# Task definition 업데이트 후 서비스 재배포
aws ecs update-service --cluster tene-prod-cluster --service tene-prod-api \
  --force-new-deployment --profile monsa-sandbox
```

또는 `git push origin main` → GitHub Actions 자동 배포

### 시크릿 변경 시

```bash
# 1. Tene vault 업데이트
tene set NEW_SECRET "value" --env prod

# 2. AWS Secrets Manager 동기화
tene run --env prod -- ./scripts/sync-secrets.sh

# 3. ECS 재배포 (새 시크릿 반영)
aws ecs update-service --cluster tene-prod-cluster --service tene-prod-api \
  --force-new-deployment --profile monsa-sandbox
```

### Dashboard (Next.js → Vercel)

- `git push origin main` → Vercel 자동 배포
- PR → Vercel Preview URL 자동 생성
- Root Directory: `apps/dashboard`
- Domain: `app.tene.sh`

### Landing Page (Next.js → Vercel)

- `git push origin main` → Vercel 자동 배포
- Root Directory: `apps/web`
- Domain: `tene.sh`

### Infrastructure (Terraform)

```bash
cd infra/terraform/environments/prod

# Plan (변경사항 미리보기)
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform plan

# Apply (주의: 프로덕션 변경)
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform apply
```

## Rollback

### API 롤백

```bash
# 이전 task definition으로 롤백
aws ecs update-service --cluster tene-prod-cluster --service tene-prod-api \
  --task-definition tene-prod-api:PREVIOUS_REVISION \
  --profile monsa-sandbox
```

### DB 롤백

```bash
# 마이그레이션 롤백 (주의)
docker exec -i tene-db psql -U tene_admin -d tene < migrations/000007_create_team_members.down.sql
```

## Monitoring

```bash
# ECS 서비스 상태
aws ecs describe-services --cluster tene-prod-cluster --services tene-prod-api \
  --profile monsa-sandbox --query "services[0].{Running:runningCount,Desired:desiredCount,Events:events[0:3]}"

# 로그 확인
aws logs tail "/ecs/tene-prod/api" --since 10m --profile monsa-sandbox

# RDS 상태
aws rds describe-db-instances --db-instance-identifier tene-prod-pg \
  --profile monsa-sandbox --query "DBInstances[0].DBInstanceStatus"
```
