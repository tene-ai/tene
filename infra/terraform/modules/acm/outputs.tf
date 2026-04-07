output "certificate_arn" { value = aws_acm_certificate.api.arn }
output "route53_zone_id" { value = data.aws_route53_zone.main.zone_id }
