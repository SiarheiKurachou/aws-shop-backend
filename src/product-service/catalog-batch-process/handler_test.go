package catalogbatchprocess

import (
	"context"
	"errors"
	"strings"
	"testing"

	"aws-shop-backend/src/product-service/core"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleCatalogBatchProcessSuccess(t *testing.T) {
	originalCreateProduct := createProduct
	t.Cleanup(func() {
		createProduct = originalCreateProduct
	})

	requests := make([]core.CreateProductRequest, 0)
	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		requests = append(requests, req)
		return core.Product{ID: "p-id"}, nil
	}

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "m1", Body: `{"title":"A","description":"d1","price":10,"count":1}`},
			{MessageId: "m2", Body: `{"title":"B","description":"d2","price":20,"count":2}`},
		},
	}

	err := HandleCatalogBatchProcess(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 createProduct calls, got: %d", len(requests))
	}

	if requests[0].Title != "A" || requests[1].Title != "B" {
		t.Fatalf("unexpected request sequence: %#v", requests)
	}
}

func TestHandleCatalogBatchProcessInvalidJSON(t *testing.T) {
	originalCreateProduct := createProduct
	t.Cleanup(func() {
		createProduct = originalCreateProduct
	})

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{}, nil
	}

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "bad-message", Body: `{"title":"A"`},
		},
	}

	err := HandleCatalogBatchProcess(context.Background(), event)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unmarshal SQS message bad-message") {
		t.Fatalf("expected unmarshal error with message id, got: %v", err)
	}
}

func TestHandleCatalogBatchProcessCreateProductError(t *testing.T) {
	originalCreateProduct := createProduct
	t.Cleanup(func() {
		createProduct = originalCreateProduct
	})

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{}, errors.New("boom")
	}

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "m-fail", Body: `{"title":"A","description":"d1","price":10,"count":1}`},
		},
	}

	err := HandleCatalogBatchProcess(context.Background(), event)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "create product from message m-fail") {
		t.Fatalf("expected create product error with message id, got: %v", err)
	}
}
