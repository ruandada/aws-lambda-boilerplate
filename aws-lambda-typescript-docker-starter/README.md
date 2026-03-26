# aws-lambda-typescript-docker-starter

A practical starter for building AWS Lambda HTTP services with:

- TypeScript
- Express app routing
- Local debug server
- Lambda adapter via `@codegenie/serverless-express`
- Lambda container image support (Docker)

## What you get

- Shared Express app (`src/app.ts`) for both local and Lambda runtime
- Promise-based Lambda handler (`src/lambda.ts`) compatible with Node.js 24 runtime behavior
- Local server entry (`src/local.ts`)
- Basic API examples:
  - `GET /`
  - `GET /health`
  - `GET /api/greet/:name?from=...`

## Prerequisites

- Node.js `v24.14.0`
- pnpm (or Corepack-enabled Node)
- Docker (optional, only needed for Lambda container image workflow)

## Quick start (local)

Install dependencies:

```bash
pnpm install
```

Start local server:

```bash
pnpm run dev
```

Default port is `3000`. You can override it:

```bash
PORT=8080 pnpm run dev
```

Verify APIs:

```bash
curl http://localhost:3000/
curl http://localhost:3000/health
curl "http://localhost:3000/api/greet/bob?from=local"
```

Example responses:

```json
{"message":"Hello World"}
{"status":"ok"}
{"message":"Hello, bob!","from":"local"}
```

## Lambda adapter verification (without Docker)

Run Lambda-focused tests:

```bash
pnpm run test:lambda
```

This validates:

- API Gateway v2 event can reach Express route and return `200`
- Non-HTTP event is rejected clearly

## Build as AWS Lambda container image

Build image:

```bash
docker build -t aws-lambda-typescript-docker-starter:local .
```

Run image locally (Lambda Runtime API):

```bash
docker run --rm -d --name aws-lambda-typescript-docker-starter -p 9000:8080 aws-lambda-typescript-docker-starter:local
```

Invoke function (API Gateway v2 style event):

```bash
curl -XPOST "http://localhost:9000/2015-03-31/functions/function/invocations" \
  -d '{"version":"2.0","routeKey":"$default","rawPath":"/","rawQueryString":"","headers":{"host":"localhost"},"requestContext":{"accountId":"123456789012","apiId":"local-api","domainName":"localhost","domainPrefix":"localhost","requestId":"request-id","routeKey":"$default","stage":"$default","time":"24/Mar/2026:00:00:00 +0000","timeEpoch":0,"http":{"method":"GET","path":"/","protocol":"HTTP/1.1","sourceIp":"127.0.0.1","userAgent":"curl"}},"isBase64Encoded":false}'
```

Expected result contains:

```json
{ "statusCode": 200, "body": "{\"message\":\"Hello World\"}" }
```

Stop container:

```bash
docker stop aws-lambda-typescript-docker-starter
```

## Deploy to AWS Lambda (ECR)

Set your variables:

```bash
export AWS_REGION="ap-southeast-1"
export AWS_ACCOUNT_ID="<your-account-id>"
export ECR_REPO="aws-lambda-typescript-docker-starter"
export IMAGE_TAG="v1"
export FUNCTION_NAME="aws-lambda-typescript-docker-starter"
```

Create ECR repository (one-time):

```bash
aws ecr create-repository --repository-name "$ECR_REPO" --region "$AWS_REGION"
```

Login Docker to ECR:

```bash
aws ecr get-login-password --region "$AWS_REGION" \
  | docker login --username AWS --password-stdin "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com"
```

Build, tag, and push image:

```bash
docker build -t "$ECR_REPO:$IMAGE_TAG" .
docker tag "$ECR_REPO:$IMAGE_TAG" "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:$IMAGE_TAG"
docker push "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:$IMAGE_TAG"
```

Create Lambda function from image (one-time):

```bash
aws lambda create-function \
  --function-name "$FUNCTION_NAME" \
  --package-type Image \
  --code ImageUri="$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:$IMAGE_TAG" \
  --role "arn:aws:iam::$AWS_ACCOUNT_ID:role/<lambda-execution-role>" \
  --region "$AWS_REGION"
```

Update existing function to new image:

```bash
aws lambda update-function-code \
  --function-name "$FUNCTION_NAME" \
  --image-uri "$AWS_ACCOUNT_ID.dkr.ecr.$AWS_REGION.amazonaws.com/$ECR_REPO:$IMAGE_TAG" \
  --region "$AWS_REGION"
```

Optional quick invoke check:

```bash
aws lambda invoke \
  --function-name "$FUNCTION_NAME" \
  --payload '{"version":"2.0","routeKey":"$default","rawPath":"/","rawQueryString":"","headers":{"host":"example.com"},"requestContext":{"accountId":"123456789012","apiId":"api-id","domainName":"example.com","domainPrefix":"example","requestId":"request-id","routeKey":"$default","stage":"$default","time":"24/Mar/2026:00:00:00 +0000","timeEpoch":0,"http":{"method":"GET","path":"/","protocol":"HTTP/1.1","sourceIp":"127.0.0.1","userAgent":"aws-cli"}},"isBase64Encoded":false}' \
  --cli-binary-format raw-in-base64-out \
  --region "$AWS_REGION" \
  response.json

cat response.json
```

## Expose via API Gateway HTTP API

If you want a public HTTPS endpoint, use API Gateway HTTP API to front the Lambda function.

Create HTTP API:

```bash
export API_NAME="aws-lambda-typescript-docker-starter-http-api"
API_ID=$(aws apigatewayv2 create-api \
  --name "$API_NAME" \
  --protocol-type HTTP \
  --region "$AWS_REGION" \
  --query 'ApiId' \
  --output text)
```

Create Lambda integration:

```bash
INTEGRATION_ID=$(aws apigatewayv2 create-integration \
  --api-id "$API_ID" \
  --integration-type AWS_PROXY \
  --integration-uri "arn:aws:lambda:$AWS_REGION:$AWS_ACCOUNT_ID:function:$FUNCTION_NAME" \
  --payload-format-version "2.0" \
  --region "$AWS_REGION" \
  --query 'IntegrationId' \
  --output text)
```

Create catch-all route:

```bash
aws apigatewayv2 create-route \
  --api-id "$API_ID" \
  --route-key "ANY /{proxy+}" \
  --target "integrations/$INTEGRATION_ID" \
  --region "$AWS_REGION"
```

Allow API Gateway to invoke Lambda:

```bash
aws lambda add-permission \
  --function-name "$FUNCTION_NAME" \
  --statement-id "apigateway-invoke-$(date +%s)" \
  --action "lambda:InvokeFunction" \
  --principal apigateway.amazonaws.com \
  --source-arn "arn:aws:execute-api:$AWS_REGION:$AWS_ACCOUNT_ID:$API_ID/*/*/*" \
  --region "$AWS_REGION"
```

Create and deploy stage:

```bash
aws apigatewayv2 create-stage \
  --api-id "$API_ID" \
  --stage-name '$default' \
  --auto-deploy \
  --region "$AWS_REGION"
```

Get invoke URL and test:

```bash
API_URL=$(aws apigatewayv2 get-api \
  --api-id "$API_ID" \
  --region "$AWS_REGION" \
  --query 'ApiEndpoint' \
  --output text)

echo "$API_URL"
curl "$API_URL/"
curl "$API_URL/health"
curl "$API_URL/api/greet/jarry?from=apigw"
```

## Build and type check

```bash
pnpm run build
pnpm run typecheck
```

## Main files

- `src/app.ts`: Express app creation + routes + async setup placeholder
- `src/local.ts`: local HTTP server bootstrap
- `src/lambda.ts`: Lambda entrypoint using `serverlessExpress({ app })`
