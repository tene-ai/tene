#!/bin/bash
# Sync secrets from Tene vault to AWS Secrets Manager
# Usage: tene run --env prod -- ./scripts/sync-secrets.sh
#
# Must be run via 'tene run' so secrets are injected as env vars.
# Tene is the source of truth; AWS Secrets Manager is the delivery mechanism for ECS.

set -euo pipefail

AWS_PROFILE="${AWS_PROFILE:-monsa-sandbox}"
SECRET_ID="tene/prod/api-secrets"

echo "  Syncing Tene vault → AWS Secrets Manager ($SECRET_ID)..."

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

echo "  ✓ Secrets synced from Tene → AWS Secrets Manager"
echo ""
echo "  To deploy: aws ecs update-service --cluster tene-prod-cluster --service tene-prod-api --force-new-deployment --profile $AWS_PROFILE"
