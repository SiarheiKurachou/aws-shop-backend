SHELL := /bin/bash

STACK_NAME ?= epam-shop-website-stack
CDK_APP ?= go run -buildvcs=false ./cdk/app
CDK ?= npx aws-cdk

.PHONY: test build-lambdas cdk-clean cdk-synth cdk-deploy build-populate-dynamo populate-dynamo populate-dynamo-append

test:
	go test ./...

build-lambdas:
	mkdir -p cdk/dist/get-products-list cdk/dist/get-product-by-id
	cp src/data.json cdk/dist/get-products-list/data.json
	cp src/data.json cdk/dist/get-product-by-id/data.json
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o cdk/dist/get-products-list/bootstrap ./cdk/lambda/get-products-list
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false -o cdk/dist/get-product-by-id/bootstrap ./cdk/lambda/get-product-by-id

cdk-clean:
	rm -rf cdk/dist && rm -rf cdk.out

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

