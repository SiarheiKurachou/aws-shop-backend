package main

import (
	"context"

	importfileparser "aws-shop-backend/src/import-service/import-file-parser"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.S3Event) error {
	return importfileparser.HandleImportFileParser(ctx, event)
}

func main() {
	lambda.Start(handler)
}
