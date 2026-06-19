SHELL := /bin/bash

STACK_NAME ?= epam-shop-website-stack
CDK_APP ?= go run -buildvcs=false ./cmd/cdk/app
CDK ?= npx aws-cdk

.PHONY: test build-lambdas build-bff-eb build-swagger cdk-clean cdk-synth cdk-deploy cdk-synth-bff cdk-deploy-bff build-populate-dynamo populate-dynamo populate-dynamo-append

test:
	go test ./...

build-lambdas:
	mkdir -p dist/get-products-list dist/get-product-by-id dist/create-product dist/import-products-file dist/import-file-parser dist/catalog-batch-process dist/basic-authorizer
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/get-products-list/bootstrap ./cmd/cdk/lambda/get-products-list
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/get-product-by-id/bootstrap ./cmd/cdk/lambda/get-product-by-id
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -tags lambda_create_product -o dist/create-product/bootstrap ./cmd/cdk/lambda/create-product
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/import-products-file/bootstrap ./cmd/cdk/lambda/import-products-file
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/import-file-parser/bootstrap ./cmd/cdk/lambda/import-file-parser
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/catalog-batch-process/bootstrap ./cmd/cdk/lambda/catalog-batch-process
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/basic-authorizer/bootstrap ./cmd/cdk/lambda/basic-authorizer

build-bff-eb:
	mkdir -p dist/bff-service-eb
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o dist/bff-service-eb/bff-service ./src/bff-service
	echo "web: ./bff-service" > dist/bff-service-eb/Procfile

build-swagger:
	@command -v swag >/dev/null 2>&1 || (echo "swag CLI not found; install with: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	mkdir -p dist/docs/product-service dist/docs/import-service dist/docs/authorization-service
	swag init -g main.go -d src/product-service -o dist/docs/product-service
	swag init -g main.go -d src/import-service -o dist/docs/import-service
	swag init -g main.go -d src/authorization-service -o dist/docs/authorization-service

cdk-clean:
	rm -rf dist && rm -rf cdk.out

cdk-synth: build-lambdas build-bff-eb
	$(CDK) synth $(STACK_NAME) --app "$(CDK_APP)"

cdk-deploy: build-lambdas build-bff-eb
	$(CDK) deploy $(STACK_NAME) --app "$(CDK_APP)" --require-approval never --outputs-file cdk.out/outputs.json

cdk-synth-bff: build-bff-eb
	$(CDK) synth bff-service-stack --app "$(CDK_APP)"

cdk-deploy-bff: build-bff-eb
	$(CDK) deploy bff-service-stack --app "$(CDK_APP)" --require-approval never --outputs-file cdk.out/bff-outputs.json

build-populate-dynamo:
	mkdir -p dist
	go build -o dist/populate-dynamo ./cmd/populate-dynamo

populate-dynamo: build-populate-dynamo
	./dist/populate-dynamo --clear

populate-dynamo-append: build-populate-dynamo
	./dist/populate-dynamo

