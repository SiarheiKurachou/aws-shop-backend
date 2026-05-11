SHELL := /bin/bash

STACK_NAME ?= epam-shop-website-stack
CDK_APP ?= go run -buildvcs=false ./cmd/cdk/app
CDK ?= npx aws-cdk

.PHONY: test build-lambdas build-swagger cdk-clean cdk-synth cdk-deploy build-populate-dynamo populate-dynamo populate-dynamo-append

test:
	go test ./...

build-lambdas:
	mkdir -p dist/get-products-list dist/get-product-by-id dist/create-product
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/get-products-list/bootstrap ./cmd/cdk/lambda/get-products-list
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/get-product-by-id/bootstrap ./cmd/cdk/lambda/get-product-by-id
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -tags lambda_create_product -o dist/create-product/bootstrap ./cmd/cdk/lambda/create-product

build-swagger:
	@command -v swag >/dev/null 2>&1 || (echo "swag CLI not found; install with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	mkdir -p dist/product-service/docs
	swag init -g main.go -d src/product-service -o dist/product-service/docs

cdk-clean:
	rm -rf dist && rm -rf cdk.out

cdk-synth: build-lambdas
	$(CDK) synth $(STACK_NAME) --app "$(CDK_APP)"

cdk-deploy: build-lambdas
	$(CDK) deploy $(STACK_NAME) --app "$(CDK_APP)" --require-approval never --outputs-file cdk.out/outputs.json

build-populate-dynamo:
	mkdir -p dist
	go build -o dist/populate-dynamo ./cmd/populate-dynamo

populate-dynamo: build-populate-dynamo
	./dist/populate-dynamo --clear

populate-dynamo-append: build-populate-dynamo
	./dist/populate-dynamo

