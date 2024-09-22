data "aws_acm_certificate" "mygo_app_cert" {
  domain   = "api.your-web-url.com"
  statuses = ["ISSUED"]
}

resource "aws_ecr_repository" "mygo_app" {
  name                 = var.image_repo_name
  image_tag_mutability = "MUTABLE"
  force_delete         = true
}

resource "aws_ecr_lifecycle_policy" "mygo_app" {
  repository = aws_ecr_repository.mygo_app.name

  policy = jsonencode({
    rules = [
      {
        rulePriority = 1
        description  = "Keep last 10 images"
        selection    = {
          tagStatus = "any"
          countType = "imageCountMoreThan"
          countNumber = 10
        }
        action = {
          type = "expire"
        }
      },
    ]
  })
}

resource "aws_ecs_cluster" "mygo_app" {
  name = var.cluster_name
}

resource "aws_ecs_task_definition" "mygo_app" {
  family                   = "mygo-app"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  execution_role_arn       = aws_iam_role.task_definition_role.arn
  container_definitions    = jsonencode([
    {
      name             = "mygo-app",
      image            = "${var.image_repo_url}:${var.image_tag}",
      essential        = true,
      portMappings     = [
        {
          containerPort = 8080,
          hostPort      = 8080
        }
      ],
      logConfiguration = {
        logDriver = "awslogs",
        options   = {
          "awslogs-group"         = aws_cloudwatch_log_group.mygo_app.name,
          "awslogs-region"        = var.aws_region,
          "awslogs-stream-prefix" = "mygo-app"
        }
      }
    }
  ])
  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "X86_64"
  }
}

resource "aws_cloudwatch_log_group" "mygo_app" {
  name = "/ecs/mygo-app"
}

resource "aws_ecs_service" "mygo_app" {
  name               = var.service_name
  cluster            = aws_ecs_cluster.mygo_app.id
  task_definition    = aws_ecs_task_definition.mygo_app.arn
  desired_count      = 1
  launch_type        = "FARGATE"
  network_configuration {
    subnets         = [aws_subnet.app_subnet_1.id, aws_subnet.app_subnet_2.id]
    security_groups = [aws_security_group.ecs_sg.id]
    assign_public_ip = true
  }
  load_balancer {
    target_group_arn = aws_lb_target_group.mygo_app.arn
    container_name   = "mygo-app"
    container_port   = 8080
  }
}

resource "aws_security_group" "ecs_sg" {
  name        = "ecs-mygo-app-sg"
  description = "Security group for ECS tasks of mygo app"
  vpc_id      = aws_vpc.app_vpc.id

  ingress {
    description = "Ingress from LB"
    from_port   = 8080
    to_port     = 8080
    protocol    = "tcp"
    security_groups = [aws_security_group.lb_sg.id]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_security_group" "lb_sg" {
  name        = "lb-mygo-app-sg"
  description = "Security group for the load balancer of mygo app"
  vpc_id      = aws_vpc.app_vpc.id

  ingress {
    description      = "Allow HTTP and HTTPS traffic from the internet"
    from_port        = 80
    to_port          = 443
    protocol         = "tcp"
    cidr_blocks      = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }

  egress {
    description = "Allow all outbound traffic"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
    ipv6_cidr_blocks = ["::/0"]
  }
}

resource "aws_lb_target_group" "mygo_app" {
  name        = "mygo-app-tg"
  port        = 8080
  protocol    = "HTTP"
  vpc_id      = aws_vpc.app_vpc.id
  target_type = "ip"
  health_check {
    path                = "/"
    interval            = 30
    timeout             = 10
    healthy_threshold   = 2
    unhealthy_threshold = 2
  }
}

resource "aws_lb" "mygo_app" {
  name               = "mygo-app-lb"
  internal           = false
  load_balancer_type = "application"
  security_groups    = [aws_security_group.lb_sg.id]
  subnets            = [aws_subnet.app_subnet_1.id, aws_subnet.app_subnet_2.id]
  enable_deletion_protection = false
  tags = {
    Name = "mygo-app-lb"
  }
}

resource "aws_lb_listener" "http_listener" {
  load_balancer_arn = aws_lb.mygo_app.arn
  port              = 80
  protocol          = "HTTP"
  default_action {
    type = "redirect"
    redirect {
      port        = "443"
      protocol    = "HTTPS"
      status_code = "HTTP_301"
    }
  }
}

resource "aws_lb_listener" "https_listener" {
  load_balancer_arn = aws_lb.mygo_app.arn
  port              = 443
  protocol          = "HTTPS"
  certificate_arn   = data.aws_acm_certificate.mygo_app_cert.arn
  ssl_policy        = "ELBSecurityPolicy-2016-08"

  default_action {
    type             = "forward"
    target_group_arn = aws_lb_target_group.mygo_app.arn
  }
}

resource "aws_iam_role" "task_definition_role" {
  name = "mygo_task_definition"
  assume_role_policy = data.aws_iam_policy_document.task_assume_role_policy.json
}

data "aws_iam_policy_document" "task_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
    effect = "Allow"
  }
}

resource "aws_iam_role_policy" "task_definition_policy" {
  name = "mygo_task_definition_policy"
  role = aws_iam_role.task_definition_role.id
  policy = data.aws_iam_policy_document.task_policy.json
}

data "aws_iam_policy_document" "task_policy" {
  statement {
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:GetAuthorizationToken",
      "ecr:GetDownloadUrlForLayer",
      "ecr:BatchGetImage",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
      "secretsmanager:GetSecretValue",
      "ssm:GetParameters",
    ]
    resources = ["*"]
    effect    = "Allow"
  }
}
