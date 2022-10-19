lambda
---

A Lambda function that simply retrieves values from the Parameter Store and outputs them to the log.

## Run at local

1. Build for local

    ```bash
    docker build -t parameters-secrets-lambda-extension:local . -f ./Dockerfile_local
    ```

2. Run at local

    ```bash
    docker run \
    --rm \
    -p 9000:8080 \
    -e AWS_DEFAULT_REGION="ap-northeast-1" \
    -e AWS_ACCESS_KEY_ID="<your-aws-access-key-id>" \
    -e AWS_SECRET_ACCESS_KEY="<your-aws-secret-access-key>"
    parameters-secrets-lambda-extension:local
    ```

3. Invoke function

    ```bash
    curl -X POST \
    -H 'Content-Type: application/json' \
    -d '{"useExtension": true, "count": 10}' \
    http://localhost:9000/2015-03-31/functions/function/invocations
    ```

## Push to ECR Repository

After creating ECR Repository, push built image to there.

0. Login to ECR

    ```bash
    REGION='ap-northeast-1'
    AWS_ACCOUNT_ID=$(
      aws sts get-caller-identity \
      --query 'Account' \
      --output text) \
    && aws ecr get-login-password \
      --region "${REGION}" \
      | docker login \
      --username AWS \
      --password-stdin ${AWS_ACCOUNT_ID}.dkr.ecr.${REGION}.amazonaws.com
    ```

1. Build for production

    ```bash
    docker build -t parameters-secrets-lambda-extension:prod .
    ```

2. Create tag

    ```bash
    docker tag \
    parameters-secrets-lambda-extension:prod \
    "${AWS_ACCOUNT_ID}".dkr.ecr.ap-northeast-1.amazonaws.com/parameters-secrets-lambda-extension:latest
    ```

3. Push

    ```bash
    docker push "${AWS_ACCOUNT_ID}".dkr.ecr.ap-northeast-1.amazonaws.com/parameters-secrets-lambda-extension:latest
    ```
    
## Invoke function

When using AWS CLI v2. (The `--cli-binary-format` option is not required when using v1.)

```bash
aws lambda invoke \
--function-name lambda-parameters-extension-function \
--invocation-type RequestResponse \
--cli-binary-format raw-in-base64-out \
--payload '{"useExtension": false, "count": 10}' \
--region ap-northeast-1 \
--log-type Tail \
/dev/stdout \
| jq -sr '.[1] | .LogResult' \
| base64 -d
```

Following output will be got.

```bash
START RequestId: c7d58c6a-e4e7-4c0b-b320-7294c6223f64 Version: $LATEST
2022/10/19 16:52:35 start handler at prod
2022/10/19 16:52:35 UseExtension: false, Count: 10
2022/10/19 16:52:35 [0] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:35 [1] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:35 [2] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:35 [3] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:35 [4] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:35 [5] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:35 [6] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:36 [7] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:36 [8] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:36 [9] extension:false value:This is test parameter key:/test/lambda-parameters-extension
2022/10/19 16:52:36 [prod] end handler
END RequestId: c7d58c6a-e4e7-4c0b-b320-7294c6223f64
REPORT RequestId: c7d58c6a-e4e7-4c0b-b320-7294c6223f64	Duration: 278.26 ms	Billed Duration: 279 ms	Memory Size: 128 MB	Max Memory Used: 26 MB
```