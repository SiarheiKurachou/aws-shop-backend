package main

import (
	"context"
	"net/http"
	"net/http/httptest"

	"aws-shop-backend/src/product_service/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(_ context.Context, _ events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	req := httptest.NewRequest(http.MethodGet, "/products", nil)
	rr := httptest.NewRecorder()
	core.HandleListProducts(rr, req)
	return events.APIGatewayProxyResponse{
		StatusCode: rr.Code,
		Headers: map[string]string{
			"Content-Type":                rr.Header().Get("Content-Type"),
			"Access-Control-Allow-Origin": rr.Header().Get("Access-Control-Allow-Origin"),
		},
		Body: rr.Body.String(),
	}, nil
}

func main() {
	lambda.Start(handler)
}
