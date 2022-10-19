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