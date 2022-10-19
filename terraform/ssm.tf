resource "aws_ssm_parameter" "sample_parameter" {
  name  = "/test/lambda-parameters-extension"
  type  = "String"
  value = "This is test parameter"

}
