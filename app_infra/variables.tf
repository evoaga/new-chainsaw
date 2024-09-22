variable "aws_region" {
  default = "<region>"
}

variable "aws_account_id" {
  description = "The AWS account ID."
  type        = string
}

variable "image_tag" {
  default = "latest"
}

variable "image_repo_name" {
  description = "The name of the ECR repository."
  type        = string
}

variable "image_repo_url" {
  default = "<account-id>.dkr.ecr.<region>.amazonaws.com/<repository>"
}

variable "github_repo_owner" {
  default = "evoaga"
}

variable "github_repo_name" {
  default = "new-chainsaw"
}

variable "github_branch" {
  default = "main"
}

variable "github_oauth_token" {
  type        = string
  description = "OAuth token for GitHub authentication"
}

variable "file_name" {
  default     = "imagedefinitions.json"
  type        = string
  description = "The file name of the image definitions."
}

variable "cluster_name" {
  type        = string
  description = "The name of the ECS Cluster."
}

variable "service_name" {
  type        = string
  description = "The name of the ECS Service."
}