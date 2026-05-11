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

# Regenerate docs.go, swagger.json, and swagger.yaml in dist/product-service/docs
make build-swagger

# If "swag" is not in PATH, run it directly:
# PATH="$(go env GOPATH)/bin:$PATH" make build-swagger
```

Swagger static files are deployed from `dist/product-service/docs/` to the website bucket under `docs/`.
After deployment, use the CloudFront domain output to access:

- `https://<cloudfront-domain>/docs/swagger.json`
- `https://<cloudfront-domain>/docs/swagger.yaml`

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
