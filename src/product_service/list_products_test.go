package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHandleListProducts(t *testing.T) {
	tests := []struct {
		name           string
		expectedStatus int
		expectedField  string
	}{
		{
			name:           "successful request",
			expectedStatus: http.StatusOK,
			expectedField:  "products",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "/products", nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(handleListProducts)
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("handler returned wrong content type: got %v want application/json", ct)
			}

			if cors := rr.Header().Get("Access-Control-Allow-Origin"); cors != "*" {
				t.Errorf("handler missing CORS header: got %v want *", cors)
			}

			var response []Product
			if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			if response == nil {
				t.Errorf("response products should not be nil")
			}

			if len(response) == 0 {
				t.Errorf("response products should not be empty")
			}
		})
	}
}

func TestLoadProducts(t *testing.T) {
	// Create a temporary test data file
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "data.json")

	testData := DataFile{
		Products: []Product{
			{
				ID:          "test-1",
				Title:       "Test Product",
				Description: "Test Description",
				Price:       19.99,
			},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		t.Fatalf("failed to write test data file: %v", err)
	}

	// Save current working directory and change to temp dir
	oldCwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get current working directory: %v", err)
	}
	defer os.Chdir(oldCwd)

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to change to temp directory: %v", err)
	}

	tests := []struct {
		name          string
		expectedCount int
		shouldFail    bool
	}{
		{
			name:          "load products successfully",
			expectedCount: 1,
			shouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create src/data.json for this test
			srcDir := filepath.Join(tmpDir, "src")
			if err := os.MkdirAll(srcDir, 0755); err != nil {
				t.Fatalf("failed to create src directory: %v", err)
			}

			srcDataFile := filepath.Join(srcDir, "data.json")
			if err := os.WriteFile(srcDataFile, data, 0644); err != nil {
				t.Fatalf("failed to write src/data.json: %v", err)
			}

			products, err := loadProducts()

			if tt.shouldFail && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.shouldFail && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.shouldFail && len(products) != tt.expectedCount {
				t.Errorf("expected %d products, got %d", tt.expectedCount, len(products))
			}

			if !tt.shouldFail && len(products) > 0 {
				if products[0].Title != "Test Product" {
					t.Errorf("expected product title 'Test Product', got '%s'", products[0].Title)
				}
				if products[0].Price != 19.99 {
					t.Errorf("expected product price 19.99, got %f", products[0].Price)
				}
			}
		})
	}
}

func TestLoadProductsWithEnvVar(t *testing.T) {
	// Create a temporary test data file
	tmpDir := t.TempDir()
	dataFile := filepath.Join(tmpDir, "custom-products.json")

	testData := DataFile{
		Products: []Product{
			{
				ID:          "env-test-1",
				Title:       "Env Test Product",
				Description: "Environment Variable Test",
				Price:       29.99,
			},
		},
	}

	data, err := json.Marshal(testData)
	if err != nil {
		t.Fatalf("failed to marshal test data: %v", err)
	}

	if err := os.WriteFile(dataFile, data, 0644); err != nil {
		t.Fatalf("failed to write test data file: %v", err)
	}

	// Save and set the environment variable
	oldEnv := os.Getenv("PRODUCTS_DATA_FILE")
	defer os.Setenv("PRODUCTS_DATA_FILE", oldEnv)

	os.Setenv("PRODUCTS_DATA_FILE", dataFile)

	products, err := loadProducts()
	if err != nil {
		t.Fatalf("failed to load products via env var: %v", err)
	}

	if len(products) != 1 {
		t.Fatalf("expected 1 product, got %d", len(products))
	}

	if products[0].ID != "env-test-1" {
		t.Errorf("expected product id 'env-test-1', got %s", products[0].ID)
	}

	if products[0].Title != "Env Test Product" {
		t.Errorf("expected title 'Env Test Product', got %s", products[0].Title)
	}
}

func TestLoadProductsEnvVarNotFound(t *testing.T) {
	// Save and set a non-existent path in the environment variable
	oldEnv := os.Getenv("PRODUCTS_DATA_FILE")
	defer os.Setenv("PRODUCTS_DATA_FILE", oldEnv)

	os.Setenv("PRODUCTS_DATA_FILE", "/tmp/non-existent-path-for-testing/data.json")

	_, err := loadProducts()
	if err == nil {
		t.Errorf("expected error when PRODUCTS_DATA_FILE path does not exist")
	}

	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("expected 'does not exist' error message, got: %v", err)
	}
}

func TestProductStruct(t *testing.T) {
	product := Product{
		ID:          "1",
		Title:       "Test",
		Description: "Desc",
		Price:       9.99,
	}

	data, err := json.Marshal(product)
	if err != nil {
		t.Fatalf("failed to marshal product: %v", err)
	}

	var unmarshaled Product
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal product: %v", err)
	}

	if unmarshaled.ID != product.ID {
		t.Errorf("product ID mismatch after marshal/unmarshal")
	}
	if unmarshaled.Title != product.Title {
		t.Errorf("product Title mismatch after marshal/unmarshal")
	}
	if unmarshaled.Description != product.Description {
		t.Errorf("product Description mismatch after marshal/unmarshal")
	}
	if unmarshaled.Price != product.Price {
		t.Errorf("product Price mismatch after marshal/unmarshal")
	}
}

func TestProductsArrayResponse(t *testing.T) {
	response := []Product{
		{
			ID:          "1",
			Title:       "Product 1",
			Description: "Desc 1",
			Price:       10.00,
		},
	}

	data, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("failed to marshal response: %v", err)
	}

	var unmarshaled []Product
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(unmarshaled) != 1 {
		t.Errorf("expected 1 product, got %d", len(unmarshaled))
	}

	if unmarshaled[0].ID != response[0].ID {
		t.Errorf("product ID mismatch after marshal/unmarshal")
	}
}
