variable "localstack_endpoint" {
  type    = string
  default = "http://localhost:4566"
}

variable "localstack_endpoint_internal" {
  type    = string
  default = "http://localstack:4566"
}

variable "pull_cache_s3_bucket" {
  type    = string
  default = "invest-source.coinmarketcap-pull-cache"
}

variable "region" {
  type    = string
  default = "us-west-2"
}

terraform {
  required_version = ">= 0.14.4"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 3.24"
    }
  }
}

provider "aws" {
  region                      = var.region
  access_key                  = "stub"
  secret_key                  = "stub"
  skip_credentials_validation = true
  skip_metadata_api_check     = true
  skip_requesting_account_id  = true
  s3_force_path_style         = true

  endpoints {
    s3               = var.localstack_endpoint
    lambda           = var.localstack_endpoint
    iam              = var.localstack_endpoint
    cloudwatch       = var.localstack_endpoint
    cloudwatchevents = var.localstack_endpoint
  }
}

locals {
  pull_lambda_name = "coinmarketcap-pull"
  pull_lambda_zip  = "${path.root}/../../../build/artifacts/coinmarketcap-pull-aws-lambda.zip"
  tags = {
    "Project" = "github.com/benjohns1/invest-source"
    "Service" = "coinmarketcap-pull-aws-lambda"
  }
  cfg = merge(try(yamldecode(file("${path.root}/../../../config.yaml")), {}), try(yamldecode(file("${path.root}/../../../.secrets.yaml")), {}))
}

resource "aws_s3_bucket" "coinmarketcap_cache" {
  bucket = var.pull_cache_s3_bucket
}

resource "aws_lambda_function" "pull_lambda" {
  function_name = local.pull_lambda_name
  role          = aws_iam_role.lambda.arn
  tags          = local.tags

  filename         = local.pull_lambda_zip
  source_code_hash = filebase64sha256(local.pull_lambda_zip)
  runtime          = "go1.x"
  handler          = "coinmarketcap-pull-aws-lambda"
  environment {
    variables = merge(local.cfg, {
      AWSEndpoint    = var.localstack_endpoint_internal
      AWSRegion      = var.region
      CacheS3Bucket  = var.pull_cache_s3_bucket
      AWS_ACCESS_KEY = "omit"
      AWS_SECRET_KEY = "omit"
    })
  }

  memory_size                    = 128
  timeout                        = 5
  reserved_concurrent_executions = 1
}

resource "aws_iam_role" "lambda" {
  name               = local.pull_lambda_name
  assume_role_policy = data.aws_iam_policy_document.lambda_role.json
  tags               = local.tags
}

data "aws_iam_policy_document" "lambda_role" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}


resource "aws_cloudwatch_event_rule" "schedule" {
  name                = local.pull_lambda_name
  schedule_expression = "rate(1 minute)"
  tags                = local.tags
}

resource "aws_cloudwatch_event_target" "schedule" {
  rule = aws_cloudwatch_event_rule.schedule.name
  arn  = aws_lambda_function.pull_lambda.arn
}

resource "aws_lambda_permission" "scheduler_permission" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.pull_lambda.function_name
  principal     = "events.amazonaws.com"
  source_arn    = aws_cloudwatch_event_rule.schedule.arn
}