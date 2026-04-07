output "vault_bucket_name" { value = aws_s3_bucket.vault.id }
output "vault_bucket_arn" { value = aws_s3_bucket.vault.arn }
output "alb_logs_bucket_name" { value = aws_s3_bucket.alb_logs.id }
