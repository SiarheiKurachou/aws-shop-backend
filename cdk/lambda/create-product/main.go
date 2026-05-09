//go:build lambda_create_product

package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"aws-shop-backend/src/product-service/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func handler(_ context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var req core.CreateProductRequest
	if err := json.Unmarshal([]byte(event.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error":"Invalid request body"}`,
		}, nil
	}

	// Create httptest request/recorder to use the handler
	httpReq := httptest.NewRequest(http.MethodPost, "/products", nil)
	rr := httptest.NewRecorder()

	// Call the handler through the request context (inject body)
	httpReq.Body = httptest.NewRequest(http.MethodPost, "/products", nil).Body
	rr = httptest.NewRecorder()

	// Manually call HandleCreateProduct
	core.HandleCreateProduct(rr, httpReq)

	// Actually, we need to pass the parsed request to create the product
	// Let's create it directly
	product, err := core.CreateProduct(req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		errMsg := err.Error()

		// Check for validation errors
		if errMsg == "title is required" ||
			errMsg == "price must be non-negative" ||
			errMsg == "count must be non-negative" {
			statusCode = http.StatusBadRequest
		}

		return events.APIGatewayProxyResponse{
			StatusCode: statusCode,
			Headers: map[string]string{
				"Content-Type":                "application/json",
				"Access-Control-Allow-Origin": "*",
			},
			Body: `{"error":"` + errMsg + `"}`,
		}, nil
	}

	// Return created product
	body, _ := json.Marshal(core.CreateProductResponse{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Count:       product.Count,
	})

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusCreated,
		Headers: map[string]string{
			"Content-Type":                "application/json",
			"Access-Control-Allow-Origin": "*",
		},
		Body: string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
