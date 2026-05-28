# aws-shop-backend
Backend for RSSchool aws-shop app including deployments

## Developent
To start locally:
```
# Load from .env file before running (bash/zsh)
set -a && source .env && set +a && go run -buildvcs=false ./src/product-service

# Or directly in the command
$(cat .env) go run -buildvcs=false ./src/product-service
```

## Import Service
The import service has two handlers:

- `import-products-file`: API Gateway `GET /import?name=<file.csv>` returns a signed S3 PUT URL for `uploaded/<file.csv>`.
- `import-file-parser`: S3 event consumer triggered by `s3:ObjectCreated:*` for objects with key prefix `uploaded/`.

When a CSV is uploaded to the signed URL, the parser lambda reads the object as a stream, treats the first row as headers, maps each next row to a key-value record, and sends each record to `catalogItemsQueue` in SQS.

### Run Locally
Run the local import-service HTTP server:

```bash
export IMPORT_BUCKET_NAME=<your-import-bucket>
go run -buildvcs=false ./src/import-service
```

The local server listens on `:8081` and exposes:

- `GET /import?name=products.csv`

Example request:

```bash
curl "http://localhost:8081/import?name=products.csv"
```

The response body is the signed URL as a plain string.

## Deploy
The CDK app deploys both the product service and the static UI from a single Go stack.

Before running CDK commands, make sure these prerequisites are in place:

- AWS credentials are configured for the target account.
- Node.js is available so `npx aws-cdk` can run.
- The frontend has already been built and the static files exist in `dist-ui/`.
- The Lambda artifacts can be built locally with `make build-lambdas`.
- Swagger static files exist in `dist/product-service/docs/`.

Export the deployment environment variables:

```bash
export AWS_REGION=eu-north-1
export AWS_ACCOUNT_ID=<your-aws-account-id>
export CDK_DEFAULT_REGION="$AWS_REGION"
export CDK_DEFAULT_ACCOUNT="$AWS_ACCOUNT_ID"

# Required for SNS email subscription on createProductTopic.
export PRODUCT_NOTIFICATION_EMAIL=<your-email@example.com>
export PRODUCT_NOTIFICATION_EMAIL_PREMIUM=<your-second-email@example.com>

# Optional: override the default website bucket name.
export S3_BUCKET_NAME=<your-unique-bucket-name>
```

Build the UI artifacts before synth or deploy. The stack reads the static site directly from `dist-ui/`, so deployment will fail if that directory is missing.

```bash
# Run this in the UI project/repo that produces the static site.
# The final build output must be copied or generated into ./dist-ui
```

Build Swagger docs before synth or deploy if API annotations changed:

```bash
# Install swag CLI (one-time setup)
go install github.com/swaggo/swag/cmd/swag@latest

# Regenerate docs.go, swagger.json, and swagger.yaml in dist/docs/*
make build-swagger

# If "swag" is not in PATH, run it directly:
# PATH="$(go env GOPATH)/bin:$PATH" make build-swagger
```

Swagger static files are deployed from `dist/docs/` to the website bucket under `docs/`.
After deployment, use the CloudFront domain output to access:

- `https://<cloudfront-domain>/docs/product-service/swagger.json`
- `https://<cloudfront-domain>/docs/product-service/swagger.yaml`
- `https://<cloudfront-domain>/docs/import-service/swagger.json`
- `https://<cloudfront-domain>/docs/import-service/swagger.yaml`
- `https://<cloudfront-domain>/docs/authorization-service/swagger.json`
- `https://<cloudfront-domain>/docs/authorization-service/swagger.yaml`

Synthesize the combined stack:

```bash
make cdk-synth
```

Deploy the combined stack:

```bash
make cdk-deploy
```

The deploy target is non-interactive and safe to run repeatedly. It always deploys the same stack and skips approval prompts (`--require-approval never`), which is useful for CI.

To deploy a different stack name (for example per environment), override `STACK_NAME`:

```bash
make cdk-deploy STACK_NAME=epam-shop-website-stack-dev
```

What gets deployed:

- API Gateway with the `GET /products` and `GET /products/{id}` endpoints
- Two Lambda functions built from `cdk/lambda/`
- An S3 bucket for the website assets
- A CloudFront distribution in front of that S3 bucket
- A bucket deployment that uploads `dist-ui/` contents and invalidates `/*`
- A bucket deployment that uploads Swagger files from `dist/product-service/docs/` to `docs/`

The stack outputs include the API URL, website bucket name, CloudFront domain name, and Lambda function names.

## Import Service Deployment Verification
Use this checklist after `make cdk-deploy`:

1. Confirm stack outputs include `products-api-url`, `imports-bucket-name`, `import-products-file-lambda-name`, and `import-file-parser-lambda-name`.
2. Generate a signed URL:
	- Run `curl "<products-api-url>import?name=test-products.csv"`.
	- Verify the response is a plain URL string.
3. Upload a CSV file with headers using the signed URL:
	- Create a sample file:

	```csv
	title,description,price,count
	Book,Example Book,10,3
	Pen,Blue Ink Pen,2,10
	Notebook,Spiral Notebook,7,5
	```

	- Example:
	  `curl -X PUT --upload-file ./test-products.csv "<signed-url-from-step-2>"`
4. Verify object location in S3:
	- Check that `test-products.csv` exists under the `uploaded/` prefix in the imports bucket.
5. Verify parser Lambda trigger and SQS delivery:
	- Open logs for `import-file-parser`.
	- Confirm invocation happened after upload.
	- Confirm messages are published to `catalogItemsQueue`.
6. Negative check (prefix filter):
	- Upload a CSV outside `uploaded/` (for example to `other/test.csv`).
	- Confirm `import-file-parser` is not triggered for that upload.

## SNS Notifications
The stack creates an SNS topic named `createProductTopic` and subscribes an email endpoint to it.

- Set `PRODUCT_NOTIFICATION_EMAIL` for the `budget` filter subscription.
- Set `PRODUCT_NOTIFICATION_EMAIL_PREMIUM` for the `premium` filter subscription.
- The `catalogBatchProcess` lambda publishes a `priceCategory` message attribute (`budget` or `premium`) so SNS filter policies route to different emails.
- After deployment, confirm both SNS subscription confirmation emails in your inbox.
