
resource "aws_iam_role" "lambdassmmiddleware" {
  name = "lambdassmmiddleware"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })
}

resource "aws_iam_policy" "lambdassmmiddleware" {
  name   = "lambdassmmiddleware"
  path   = "/"
  policy = data.aws_iam_policy_document.lambdassmmiddleware.json
}

resource "aws_iam_role_policy_attachment" "lambdassmmiddleware" {
  role       = aws_iam_role.lambdassmmiddleware.name
  policy_arn = aws_iam_policy.lambdassmmiddleware.arn
}

data "aws_iam_policy_document" "lambdassmmiddleware" {
  statement {
    actions = [
      "ssm:GetParameter*",
      "ssm:DescribeParameters",
      "ssm:List*",
    ]
    resources = ["*"]
  }
  statement {
    actions = [
      "logs:GetLog*",
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]
    resources = ["*"]
  }
}

resource "aws_ssm_parameter" "foo" {
  name        = "/lambdassmmiddleware/foo"
  description = "foo for lambdassmmiddleware"
  type        = "String"
  value       = "foo values"
}

resource "aws_ssm_parameter" "bar" {
  name        = "/lambdassmmiddleware/bar"
  description = "bar for lambdassmmiddleware"
  type        = "String"
  value       = "bar values"
}

resource "aws_ssm_parameter" "hoge" {
  name        = "/lambdassmmiddleware/paths/hoge"
  description = "hoge for lambdassmmiddleware"
  type        = "String"
  value       = "hoge values"
}
