resource "aws_codepipeline" "codepipeline" {
  name     = "mygo_pipeline"
  role_arn = aws_iam_role.codepipeline_role.arn
  artifact_store {
    location = aws_s3_bucket.codepipeline_bucket.bucket
    type     = "S3"
    encryption_key {
      id   = aws_kms_key.mygo_app_s3_kms_key.arn
      type = "KMS"
    }
  }
  stage {
    name = "Source"
    action {
      name             = "SourceAction"
      
      category         = "Source"
      owner            = "ThirdParty"
      provider         = "GitHub"
      version          = "1"
      output_artifacts = ["source_output"]
      configuration = {
        Owner      = var.github_repo_owner
        Repo       = var.github_repo_name
        Branch     = var.github_branch
        OAuthToken = var.github_oauth_token
      }
    }
  }
  stage {
    name = "Build"
    action {
      name             = "Build"

      category         = "Build"
      owner            = "AWS"
      provider         = "CodeBuild"
      input_artifacts  = ["source_output"]
      output_artifacts = ["build_output"]
      version          = "1"
      configuration = {
        ProjectName = aws_codebuild_project.mygo_app.name
      }
    }
  }

  stage {
    name = "Deploy"

    action {
      name            = "Deploy"
      category        = "Deploy"
      owner           = "AWS"
      provider        = "ECS"
      version         = 1
      run_order       = 1
      input_artifacts = ["build_output"]

      configuration = {
        ClusterName = var.cluster_name
        ServiceName = var.service_name
        FileName = "${var.file_name}"
      }
    }
  }
}

resource "aws_s3_bucket" "codepipeline_bucket" {
  bucket = "mygo-demo-codepipeline-bucket-evoaga"
  force_destroy = true
}

resource "aws_iam_role" "codepipeline_role" {
  name               = "mygo_codepipeline_role"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Sid       = "TrustPolicyStatementThatAllowsEC2ServiceToAssumeTheAttachedRole"
        Effect    = "Allow"
        Principal = {
          Service = "codepipeline.amazonaws.com"
        }
        Action    = "sts:AssumeRole"
      },
    ]
  })
}

resource "aws_iam_role_policy" "codepipeline_policy" {
  name = "mygo_codepipeline_policy"
  role = aws_iam_role.codepipeline_role.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:GetBucketVersioning",
          "s3:PutObjectAcl",
          "s3:PutObject",
          "iam:GetRole",
          "iam:PassRole",
        ]
        Resource = [
          aws_s3_bucket.codepipeline_bucket.arn,
          "${aws_s3_bucket.codepipeline_bucket.arn}/*",
        ]
      },
      {
        Effect   = "Allow"
        Action   = [
          "codebuild:BatchGetBuilds",
          "codebuild:StartBuild",
        ]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = [
          "cloudformation:DescribeStacks",
          "kms:GenerateDataKey",
          "iam:GetRole",
          "iam:PassRole",
        ]
        Resource = "*"
      },
      {
        Effect   = "Allow"
        Action   = "kms:Decrypt"
        Resource = aws_kms_key.mygo_app_s3_kms_key.arn
      },
      {
        Effect = "Allow"
        Action = [
          "ecs:DescribeServices",
          "ecs:DescribeTaskDefinition",
          "ecs:DescribeTasks",
          "ecs:ListTasks",
          "ecs:RegisterTaskDefinition",
          "ecs:UpdateService",
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_codebuild_project" "mygo_app" {
  name          = "mygo-app"
  description   = "Builds a mygo application"
  service_role  = aws_iam_role.codebuild_role.arn
  build_timeout = "5"
  artifacts {
    type = "CODEPIPELINE"
  }
  environment {
    compute_type    = "BUILD_GENERAL1_SMALL"
    image           = "aws/codebuild/amazonlinux2-x86_64-standard:4.0"
    type            = "LINUX_CONTAINER"
    privileged_mode = true
    environment_variable {
      name  = "AWS_ACCOUNT_ID"
      value = var.aws_account_id
    }
    environment_variable {
      name  = "AWS_DEFAULT_REGION"
      value = var.aws_region
    }
    environment_variable {
      name  = "IMAGE_REPO_NAME"
      value = var.image_repo_name
    }
    environment_variable {
      name  = "ECS_CLUSTER_NAME"
      value = aws_ecs_cluster.mygo_app.name
    }
    environment_variable {
      name  = "ECS_SERVICE_NAME"
      value = aws_ecs_service.mygo_app.name
    }
    environment_variable {
      name  = "ECS_TASK_DEFINITION"
      value = aws_ecs_task_definition.mygo_app.family
    }
  }
  source {
    type      = "CODEPIPELINE"
    buildspec = "buildspec.yml"
  }
}

resource "aws_iam_role" "codebuild_role" {
  name               = "mygo_codebuild_role"
  assume_role_policy = jsonencode({
    Version   = "2012-10-17"
    Statement = [
      {
        Action    = "sts:AssumeRole"
        Principal = {
          Service = "codebuild.amazonaws.com"
        }
        Effect    = "Allow"
        Sid       = "TrustPolicyStatementThatAllowsEC2ServiceToAssumeTheAttachedRole"
      },
    ]
  })
}

resource "aws_iam_role_policy" "codebuild_policy" {
  name = "mygo_codebuild_policy"
  role = aws_iam_role.codebuild_role.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ]
        Resource = "arn:aws:logs:*:*:*"
        Effect   = "Allow"
      },
      {
        Action = [
          "s3:GetObject",
          "s3:GetObjectVersion",
          "s3:GetBucketVersioning",
          "s3:PutObject"
        ]
        Resource = [
          "arn:aws:s3:::${aws_s3_bucket.codepipeline_bucket.bucket}",
          "arn:aws:s3:::${aws_s3_bucket.codepipeline_bucket.bucket}/*"
        ]
        Effect   = "Allow"
      },
      {
        Action = [
          "ecr:BatchCheckLayerAvailability",
          "ecr:CompleteLayerUpload",
          "ecr:GetAuthorizationToken",
          "ecr:InitiateLayerUpload",
          "ecr:PutImage",
          "ecr:UploadLayerPart"
        ]
        Resource = ["*"]
        Effect   = "Allow"
      },
      {
        Action = [
          "ecs:UpdateService",
          "iam:GetRole",
          "iam:PassRole"
        ]
        Resource = ["*"]
        Effect   = "Allow"
      },
      {
        Effect   = "Allow"
        Action   = [
          "kms:Decrypt",
          "kms:GenerateDataKey"
        ]
        Resource = aws_kms_key.mygo_app_s3_kms_key.arn
      },
      {
        Effect = "Allow"
        Action = [
          "codebuild:CreateReportGroup",
          "codebuild:CreateReport",
          "codebuild:BatchPutTestCases",
          "codebuild:UpdateReport",
          "codebuild:BatchGetReports"
        ]
        Resource = "*"
      } 
    ]
  })
}
