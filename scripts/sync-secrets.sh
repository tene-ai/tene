#!/bin/bash
# Sync secrets from Tene vault to AWS Secrets Manager
# Usage:
#   tene run --env prod    -- ./scripts/sync-secrets.sh prod
#   tene run --env staging -- ./scripts/sync-secrets.sh staging
#
# Must be run via 'tene run' so secrets are injected as env vars.
# Tene is the source of truth; AWS Secrets Manager is the delivery mechanism for ECS.

set -euo pipefail

# Determine environment from argument or ENV variable
TARGET_ENV="${1:-${ENV:-prod}}"
AWS_PROFILE="${AWS_PROFILE:-monsa-sandbox}"
SECRET_ID="tene/${TARGET_ENV}/api-secrets"

echo "  Syncing Tene vault (${TARGET_ENV}) → AWS Secrets Manager ($SECRET_ID)..."

SECRET_JSON=$(cat <<EOF
{
  "db_password": "${DB_PASSWORD}",
  "jwt_secret": "${JWT_SECRET}",
  "github_client_id": "${GITHUB_CLIENT_ID}",
  "github_client_secret": "${GITHUB_CLIENT_SECRET}",
  "lemon_api_key": "${LEMON_API_KEY}",
  "lemon_webhook_secret": "${LEMON_WEBHOOK_SECRET}"
}
EOF
)

aws secretsmanager put-secret-value \
  --secret-id "$SECRET_ID" \
  --secret-string "$SECRET_JSON" \
  --profile "$AWS_PROFILE" \
  --output text --query "Name"

echo "  ✓ Secrets synced from Tene (${TARGET_ENV}) → AWS Secrets Manager"
echo ""

if [ "$TARGET_ENV" = "prod" ]; then
  echo "  To deploy: aws ecs update-service --cluster tene-prod-cluster --service tene-prod-api --force-new-deployment --profile $AWS_PROFILE"
elif [ "$TARGET_ENV" = "staging" ]; then
  echo "  To deploy: aws ecs update-service --cluster tene-staging-cluster --service tene-staging-api --force-new-deployment --profile $AWS_PROFILE"
fi
