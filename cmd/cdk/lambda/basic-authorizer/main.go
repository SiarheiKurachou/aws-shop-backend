package main

import (
	"context"

	basicauthorizer "aws-shop-backend/src/authorization-service/basic-authorizer"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	return basicauthorizer.HandleBasicAuthorizer(ctx, event)
}

func main() {
	lambda.Start(handler)
}
