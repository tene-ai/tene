output "alarm_arns" {
  value = concat(
    [aws_cloudwatch_metric_alarm.ecs_cpu_high.arn],
    [aws_cloudwatch_metric_alarm.ecs_memory_high.arn],
    [aws_cloudwatch_metric_alarm.ecs_running_count.arn],
    aws_cloudwatch_metric_alarm.rds_cpu_high[*].arn,
    aws_cloudwatch_metric_alarm.rds_free_storage[*].arn,
    aws_cloudwatch_metric_alarm.alb_5xx[*].arn,
    aws_cloudwatch_metric_alarm.alb_unhealthy_hosts[*].arn,
  )
}
