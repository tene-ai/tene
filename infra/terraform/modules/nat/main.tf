# fck-nat: cost-optimized NAT ($3/mo vs NAT Gateway $32/mo)

data "aws_ami" "fck_nat" {
  most_recent = true
  owners      = ["568608671756"]

  filter {
    name   = "name"
    values = ["fck-nat-al2023-*-arm64-*"]
  }
  filter {
    name   = "architecture"
    values = ["arm64"]
  }
}

resource "aws_security_group" "nat" {
  name_prefix = "${var.project}-${var.environment}-nat-"
  vpc_id      = var.vpc_id

  ingress {
    description = "All from VPC"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = { Name = "${var.project}-${var.environment}-nat-sg" }
}

resource "aws_network_interface" "nat" {
  subnet_id         = var.public_subnet_id
  security_groups   = [aws_security_group.nat.id]
  source_dest_check = false

  tags = { Name = "${var.project}-${var.environment}-nat-eni" }
}

resource "aws_eip" "nat" {
  domain            = "vpc"
  network_interface = aws_network_interface.nat.id

  tags = { Name = "${var.project}-${var.environment}-nat-eip" }
}

resource "aws_instance" "nat" {
  ami           = data.aws_ami.fck_nat.id
  instance_type = "t4g.nano"

  network_interface {
    network_interface_id = aws_network_interface.nat.id
    device_index         = 0
  }

  tags = { Name = "${var.project}-${var.environment}-fck-nat" }

  lifecycle {
    ignore_changes = [ami]
  }
}

# Auto recovery
resource "aws_cloudwatch_metric_alarm" "nat_recovery" {
  alarm_name          = "${var.project}-${var.environment}-nat-recovery"
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = 2
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = 60
  statistic           = "Maximum"
  threshold           = 0

  dimensions = {
    InstanceId = aws_instance.nat.id
  }

  alarm_actions = ["arn:aws:automate:${var.aws_region}:ec2:recover"]
}

# Route from private subnets through NAT
resource "aws_route" "private_nat" {
  route_table_id         = var.private_route_table_id
  destination_cidr_block = "0.0.0.0/0"
  network_interface_id   = aws_network_interface.nat.id
}
