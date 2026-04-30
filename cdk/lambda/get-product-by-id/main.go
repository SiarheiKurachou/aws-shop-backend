package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"

	"aws-shop-backend/src/product-service/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(_ context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	productID := event.PathParameters["id"]
	if productID == "" {
		productID = event.QueryStringParameters["id"]
	}

	path := fmt.Sprintf("/products/%s", productID)
	req := httptest.NewRequest(http.MethodGet, path, nil)
	rr := httptest.NewRecorder()
	core.HandleGetProductByID(rr, req)
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
