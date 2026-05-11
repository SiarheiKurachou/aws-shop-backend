package main

import (
	"context"

	importproductsfile "aws-shop-backend/src/import-service/import-products-file"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	return importproductsfile.HandleImportProductsFile(ctx, event)
}

func main() {
	lambda.Start(handler)
}
