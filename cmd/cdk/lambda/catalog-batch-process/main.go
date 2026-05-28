package main

import (
	"context"

	catalogbatchprocess "aws-shop-backend/src/product-service/catalog-batch-process"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.SQSEvent) error {
	return catalogbatchprocess.HandleCatalogBatchProcess(ctx, event)
}

func main() {
	lambda.Start(handler)
}
