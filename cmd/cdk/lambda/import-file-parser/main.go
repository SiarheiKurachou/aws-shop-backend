package main

import (
	"context"

	importservice "aws-shop-backend/src/import-service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.S3Event) error {
	return importservice.HandleImportFileParser(ctx, event)
}

func main() {
	lambda.Start(handler)
}
