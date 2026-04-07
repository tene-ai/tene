resource "aws_secretsmanager_secret" "api" {
  name                    = "${var.project}/${var.environment}/api-secrets"
  description             = "API secrets for ${var.project} ${var.environment}"
  recovery_window_in_days = 30

  tags = { Name = "${var.project}-${var.environment}-api-secrets" }
}

resource "aws_secretsmanager_secret_version" "api" {
  secret_id = aws_secretsmanager_secret.api.id
  secret_string = jsonencode({
    db_password    = var.db_password
    jwt_secret     = var.jwt_secret
    lemon_api_key        = var.lemon_api_key
    lemon_webhook_secret = var.lemon_webhook_secret
    github_client_id     = var.github_client_id
    github_client_secret = var.github_client_secret
  })

  lifecycle {
    ignore_changes = [secret_string]
  }
}
