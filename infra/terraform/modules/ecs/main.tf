resource "aws_ecs_cluster" "this" {
  name = "${var.project}-${var.environment}-cluster"

  setting {
    name  = "containerInsights"
    value = "enabled"
  }

  tags = { Name = "${var.project}-${var.environment}-cluster" }
}

resource "aws_cloudwatch_log_group" "app" {
  name              = "/ecs/${var.project}-${var.environment}/api"
  retention_in_days = 30
}

resource "aws_security_group" "ecs" {
  name_prefix = "${var.project}-${var.environment}-ecs-"
  vpc_id      = var.vpc_id

  ingress {
    description     = "From ALB"
    from_port       = var.container_port
    to_port         = var.container_port
    protocol        = "tcp"
    security_groups = [var.alb_security_group_id]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.project}-${var.environment}-ecs-sg" }
}

resource "aws_ecs_task_definition" "api" {
  family                   = "${var.project}-${var.environment}-api"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = var.execution_role_arn
  task_role_arn            = var.task_role_arn

  container_definitions = jsonencode([{
    name      = "tene-api"
    image     = "${var.ecr_repository_url}:${var.image_tag}"
    essential = true

    portMappings = [{
      containerPort = var.container_port
      protocol      = "tcp"
    }]

    healthCheck = {
      command     = ["CMD-SHELL", "wget -qO- http://localhost:${var.container_port}/health || exit 1"]
      interval    = 30
      timeout     = 5
      retries     = 3
      startPeriod = 60
    }

    environment = [
      { name = "ENV", value = var.environment },
      { name = "PORT", value = tostring(var.container_port) },
      { name = "DB_HOST", value = var.db_host },
      { name = "DB_PORT", value = "5432" },
      { name = "DB_NAME", value = var.db_name },
      { name = "S3_BUCKET", value = var.s3_bucket_name },
      { name = "AWS_REGION", value = var.aws_region },
      { name = "CALLBACK_BASE", value = "https://api.tene.sh" },
      { name = "LEMON_STORE_ID", value = var.lemon_store_id },
      { name = "LEMON_VARIANT_PRO", value = var.lemon_variant_pro },
    ]

    secrets = [
      { name = "DB_PASSWORD", valueFrom = "${var.secrets_arn}:db_password::" },
      { name = "JWT_SECRET", valueFrom = "${var.secrets_arn}:jwt_secret::" },
      { name = "GITHUB_CLIENT_ID", valueFrom = "${var.secrets_arn}:github_client_id::" },
      { name = "GITHUB_CLIENT_SECRET", valueFrom = "${var.secrets_arn}:github_client_secret::" },
      { name = "LEMON_API_KEY", valueFrom = "${var.secrets_arn}:lemon_api_key::" },
      { name = "LEMON_WEBHOOK_SECRET", valueFrom = "${var.secrets_arn}:lemon_webhook_secret::" },
    ]

    logConfiguration = {
      logDriver = "awslogs"
      options = {
        "awslogs-group"         = aws_cloudwatch_log_group.app.name
        "awslogs-region"        = var.aws_region
        "awslogs-stream-prefix" = "api"
      }
    }
  }])
}

resource "aws_ecs_service" "api" {
  name            = "${var.project}-${var.environment}-api"
  cluster         = aws_ecs_cluster.this.id
  task_definition = aws_ecs_task_definition.api.arn
  desired_count   = var.desired_count
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = var.private_subnet_ids
    security_groups  = [aws_security_group.ecs.id]
    assign_public_ip = false
  }

  load_balancer {
    target_group_arn = var.target_group_arn
    container_name   = "tene-api"
    container_port   = var.container_port
  }

  deployment_minimum_healthy_percent = 100
  deployment_maximum_percent         = 200

  deployment_circuit_breaker {
    enable   = true
    rollback = true
  }

  lifecycle {
    ignore_changes = [desired_count, task_definition]
  }
}

# Auto Scaling
resource "aws_appautoscaling_target" "ecs" {
  max_capacity       = var.max_capacity
  min_capacity       = var.min_capacity
  resource_id        = "service/${aws_ecs_cluster.this.name}/${aws_ecs_service.api.name}"
  scalable_dimension = "ecs:service:DesiredCount"
  service_namespace  = "ecs"
}

resource "aws_appautoscaling_policy" "cpu" {
  name               = "${var.project}-${var.environment}-cpu-scaling"
  policy_type        = "TargetTrackingScaling"
  resource_id        = aws_appautoscaling_target.ecs.resource_id
  scalable_dimension = aws_appautoscaling_target.ecs.scalable_dimension
  service_namespace  = aws_appautoscaling_target.ecs.service_namespace

  target_tracking_scaling_policy_configuration {
    predefined_metric_specification {
      predefined_metric_type = "ECSServiceAverageCPUUtilization"
    }
    target_value       = 70.0
    scale_in_cooldown  = 300
    scale_out_cooldown = 60
  }
}
