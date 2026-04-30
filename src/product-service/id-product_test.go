package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleGetProductByIDSuccess(t *testing.T) {
	products, err := loadProducts()
	if err != nil {
		t.Fatalf("failed to load products for test setup: %v", err)
	}
	if len(products) == 0 {
		t.Fatal("expected at least one product in data.json")
	}

	targetID := products[0].ID

	req := httptest.NewRequest(http.MethodGet, "/products/"+targetID, nil)
	rr := httptest.NewRecorder()

	handleGetProductByID(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}

	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("expected content-type application/json, got %s", ct)
	}

	if cors := rr.Header().Get("Access-Control-Allow-Origin"); cors != "*" {
		t.Fatalf("expected CORS header '*', got %s", cors)
	}

	var got Product
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode product response: %v", err)
	}

	if got.ID != targetID {
		t.Fatalf("expected product id %s, got %s", targetID, got.ID)
	}
}

func TestHandleGetProductByIDNotFound(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/products/non-existent-id", nil)
	rr := httptest.NewRecorder()

	handleGetProductByID(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	var got map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if got["error"] != "Product not found" {
		t.Fatalf("expected error 'Product not found', got %q", got["error"])
	}
}

func TestHandleGetProductByIDMissingID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/products/", nil)
	rr := httptest.NewRecorder()

	handleGetProductByID(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}

	var got map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&got); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if got["error"] != "Product id is required" {
		t.Fatalf("expected error 'Product id is required', got %q", got["error"])
	}
}
