output "alb_arn" { value = aws_lb.this.arn }
output "alb_arn_suffix" { value = aws_lb.this.arn_suffix }
output "alb_dns_name" { value = aws_lb.this.dns_name }
output "alb_zone_id" { value = aws_lb.this.zone_id }
output "target_group_arn" { value = aws_lb_target_group.ecs.arn }
output "target_group_arn_suffix" { value = aws_lb_target_group.ecs.arn_suffix }
output "alb_security_group_id" { value = aws_security_group.alb.id }
