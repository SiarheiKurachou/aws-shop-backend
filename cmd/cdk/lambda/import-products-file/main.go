package main

import (
	"context"

	importservice "aws-shop-backend/src/import-service"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return importservice.HandleImportProductsFile(ctx, event)
}

func main() {
	lambda.Start(handler)
}
