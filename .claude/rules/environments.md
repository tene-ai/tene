# Tene Environments & Infrastructure

## Environment Overview

| Environment | Purpose | Secret Source |
|-------------|---------|---------------|
| **local** | 개발/디버깅 | `tene run --env local` |
| **staging** | QA/통합 테스트 | `tene run --env staging` → Secrets Manager |
| **prod** | 프로덕션 서비스 | `tene run --env prod` → Secrets Manager → ECS |

---

## Local Environment

### Services & Ports

| Port | Service | Technology | 실행 방법 |
|:----:|---------|------------|----------|
| 3000 | Landing Page | Next.js 15 (apps/web) | `npx next dev` |
| 3001 | Dashboard | Next.js 15 (apps/dashboard) | `npx next dev --port 3001` |
| 5432 | PostgreSQL | Docker (postgres:16-alpine) | `docker compose up` |
| 8080 | Go API Server | Go 1.25 + Echo v4 (cmd/server) | `tene run --env local -- go run ./cmd/server` |
| 9000 | MinIO (S3 호환) | Docker (minio/minio) | `docker compose up` |
| 9001 | MinIO Console | Docker | 브라우저: minioadmin/minioadmin |

### 시크릿 주입

```bash
# API 서버 — tene이 local 환경 시크릿을 환경변수로 주입
tene run --env local -- go run ./cmd/server

# 주입되는 변수 목록 확인
tene list --env local
```

### 실행 명령어

```bash
# 전체 시작 (Docker 인프라 + 모든 앱 서버)
./scripts/dev.sh

# 개별 실행
./scripts/dev.sh infra        # PostgreSQL + MinIO만
./scripts/dev.sh api          # Go API만 (infra 필요)
./scripts/dev.sh dashboard    # Dashboard만
./scripts/dev.sh landing      # Landing만

# 상태 확인 / 중지
./scripts/dev.sh status
./scripts/dev.sh stop
```

### 파일 구성

```
docker-compose.dev.yml    # PostgreSQL + MinIO 컨테이너
scripts/dev.sh            # 로컬 개발 오케스트레이터
/tmp/tene-api.log         # API 서버 로그
/tmp/tene-dashboard.log   # Dashboard 로그
/tmp/tene-landing.log     # Landing 로그
```

---

## Staging Environment

### 인프라

| Service | Technology | Endpoint |
|---------|------------|----------|
| API | ECS Fargate | (staging ALB) |
| DB | RDS PostgreSQL 16 | (staging RDS endpoint) |
| S3 | AWS S3 | tene-vault-staging |
| Frontend | Vercel Preview | PR별 자동 생성 |

### 배포

- PR 생성 시 Vercel Preview 자동 배포
- staging 브랜치 push 시 ECS staging 배포 (CI/CD)

---

## Production Environment

### 인프라 (AWS ap-northeast-2, Account: 507221376909 monsa-sandbox)

| Service | Technology | Endpoint | 월 비용 |
|---------|------------|----------|:------:|
| Landing | Vercel | `tene.sh` | $0 |
| Dashboard | Vercel | `app.tene.sh` | $0 |
| API Server | ECS Fargate (0.25 vCPU, 512 MiB) | `api.tene.sh` | ~$9 |
| Load Balancer | ALB + ACM (HTTPS) | ALB → ECS | ~$28 |
| Database | RDS PostgreSQL 16.6 (db.t4g.micro) | private subnet | ~$12 |
| Object Storage | S3 (SSE-S3 AES-256) | `tene-vault-prod` | ~$0.05 |
| NAT | fck-nat (t4g.nano) | private → internet | ~$3 |
| DNS | Route 53 | `tene.sh` 호스팅 존 | ~$2 |
| Secrets | Secrets Manager | `tene/prod/api-secrets` | ~$0.50 |
| Container Registry | ECR | `tene-api` | ~$0.10 |
| **합계** | | | **~$55/월** |

### 네트워크 구성

```
VPC: 10.0.0.0/16 (ap-northeast-2)
  Public:   10.0.1.0/24, 10.0.2.0/24   (ALB, fck-nat)
  Private:  10.0.10.0/24, 10.0.11.0/24 (ECS Fargate)
  Isolated: 10.0.20.0/24, 10.0.21.0/24 (RDS)

Security Groups:
  ALB:  443/tcp from 0.0.0.0/0
  ECS:  8080/tcp from ALB only
  RDS:  5432/tcp from ECS only
```

### 태그 정책

모든 AWS 리소스에 다음 태그가 자동 적용됨 (Terraform default_tags):

| Key | Value |
|-----|-------|
| Project | tene |
| Environment | prod |
| ManagedBy | terraform |

---

## CI/CD Pipeline

### Git Branch Strategy

```
main ─────────────────────────────────────────► prod 자동 배포
  │
  ├── feature/xxx ──► PR ──► Vercel Preview
  │                         + CI (test/lint)
  │                         + merge → main → deploy
  │
  └── hotfix/xxx ───► PR ──► 긴급 merge → main → deploy
```

| 브랜치 | 용도 | 배포 |
|--------|------|------|
| `main` | 프로덕션 | push 시 자동 배포 (API → ECS, Frontend → Vercel) |
| `feature/*` | 기능 개발 | PR → Vercel Preview + CI 검증 |
| `hotfix/*` | 긴급 수정 | PR → fast merge → main → 자동 배포 |

### Go API 배포 (GitHub Actions → ECR → ECS)

```
Push to main
  → Test + Lint (병렬)
  → Docker build (linux/amd64)
  → ECR push (OIDC, 장기 크레덴셜 없음)
  → ECS task definition 업데이트
  → ECS service force-new-deployment
  → Rolling update (100% → 200% → 100%)
```

워크플로우: `.github/workflows/ci.yml`

### Frontend 배포 (Vercel Git Integration)

```
Push to main → Vercel 자동 production 배포
PR 생성       → Vercel Preview 자동 생성
```

| 프로젝트 | Root Directory | Domain |
|---------|----------------|--------|
| tene (Landing) | `apps/web` | `tene.sh` |
| tene-dashboard | `apps/dashboard` | `app.tene.sh` |

워크플로우: `.github/workflows/dashboard.yml` (type check + build 검증)

---

## Secret Management Flow

```
tene vault (source of truth)
    │
    ├── local:   tene run --env local -- go run ./cmd/server
    │            (직접 환경변수 주입, Docker 인프라와 연동)
    │
    ├── staging: tene run --env staging -- ./scripts/sync-secrets.sh
    │            (Tene → Secrets Manager → ECS)
    │
    └── prod:    tene run --env prod -- ./scripts/sync-secrets.sh
                 (Tene → Secrets Manager → ECS)
```

### 시크릿 동기화 (Tene → AWS)

```bash
# Tene vault의 prod 시크릿을 AWS Secrets Manager로 동기화
tene run --env prod -- ./scripts/sync-secrets.sh

# 동기화 후 ECS 재배포
aws ecs update-service --cluster tene-prod-cluster --service tene-prod-api \
  --force-new-deployment --profile monsa-sandbox
```

---

## Terraform

### 구조

```
infra/terraform/
├── modules/          # 재사용 모듈 (12개)
│   ├── vpc/          VPC, 서브넷, IGW, 라우팅
│   ├── nat/          fck-nat (t4g.nano, NAT GW 대체)
│   ├── ecs/          Cluster, Task Def, Service, Auto Scaling
│   ├── alb/          ALB, HTTPS 리스너, Target Group
│   ├── rds/          PostgreSQL, 파라미터 그룹, 보안 그룹
│   ├── s3/           Vault 버킷, ALB 로그 버킷
│   ├── ecr/          컨테이너 레지스트리
│   ├── route53/      DNS A 레코드 (ALB alias)
│   ├── acm/          TLS 인증서 (DNS 검증)
│   ├── iam/          ECS Role, GitHub Actions OIDC
│   ├── secrets/      Secrets Manager
│   └── (waf/)        향후 추가
└── environments/
    └── prod/         프로덕션 환경 (monsa-sandbox account)
```

### 명령어

```bash
cd infra/terraform/environments/prod

# Plan
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform plan

# Apply
TF_VAR_db_password="$(tene get DB_PASSWORD --env prod)" \
TF_VAR_jwt_secret="$(tene get JWT_SECRET --env prod)" \
terraform apply
```

---

## Quick Reference

| 작업 | 명령어 |
|------|--------|
| 로컬 전체 시작 | `./scripts/dev.sh` |
| 로컬 전체 중지 | `./scripts/dev.sh stop` |
| 로컬 상태 확인 | `./scripts/dev.sh status` |
| 시크릿 목록 | `tene list --env local` |
| 시크릿 추가 | `tene set KEY VALUE --env local` |
| prod 시크릿 → AWS 동기화 | `tene run --env prod -- ./scripts/sync-secrets.sh` |
| prod 배포 (수동) | `docker build + ECR push + ECS update-service` |
| prod 배포 (자동) | `git push origin main` |
| Terraform plan | `cd infra/terraform/environments/prod && terraform plan` |
| API 헬스체크 | `curl https://api.tene.sh/health` |
