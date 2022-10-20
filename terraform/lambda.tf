resource "aws_lambda_function" "function" {
  function_name = "lambda-parameters-extension-function"
  role          = aws_iam_role.role_for_lambda.arn
  package_type  = "Image"
  image_uri     = "${aws_ecr_repository.lambda_image_repository.repository_url}:latest"
  timeout       = 60
  memory_size   = 128
  environment {
    variables = {
      "ENV" = "prod"
    }
  }
}

resource "aws_iam_role" "role_for_lambda" {
  name = "lambda-parameters-extension-function-role"

  assume_role_policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Action" : "sts:AssumeRole",
        "Principal" : {
          "Service" : "lambda.amazonaws.com"
        },
        "Effect" : "Allow",
        "Sid" : ""
      }
    ]
  })

}

resource "aws_iam_policy" "policy_for_function" {
  name = "lambda-parameters-extension-function-policy"
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Effect" : "Allow",
        "Action" : [
          "logs:CreateLogGroup",
          "logs:CreateLogStream",
          "logs:PutLogEvents"
        ],
        "Resource" : "*"
      },
      {
        "Effect" : "Allow",
        "Action" : [
          "ssm:GetParameter"
        ],
        "Resource" : "${aws_ssm_parameter.sample_parameter.arn}"
      }
    ]
  })
}

resource "aws_iam_policy_attachment" "policy_attachment_for_function_role" {
  name = "policy-attachment-for-lambda-parameters-extension-function-role"

  roles = [
    aws_iam_role.role_for_lambda.name
  ]

  policy_arn = aws_iam_policy.policy_for_function.arn
}

