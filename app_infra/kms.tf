resource "aws_kms_key" "mygo_app_s3_kms_key" {
  description             = "KMS key for encrypting CodePipeline artifacts in S3"
  deletion_window_in_days = 10
}

resource "aws_kms_alias" "mygo_app_s3_kms_key_alias" {
  name          = "alias/mygo_app_s3kmskey"
  target_key_id = aws_kms_key.mygo_app_s3_kms_key.key_id
}