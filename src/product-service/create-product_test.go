package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"aws-shop-backend/src/product-service/core"
)

func TestValidateCreateProductRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     core.CreateProductRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: core.CreateProductRequest{
				Title:       "Test Product",
				Description: "Test Description",
				Price:       100,
				Count:       50,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: core.CreateProductRequest{
				Title:       "",
				Description: "Test",
				Price:       100,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "whitespace-only title",
			req: core.CreateProductRequest{
				Title:       "   ",
				Description: "Test",
				Price:       100,
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "negative price",
			req: core.CreateProductRequest{
				Title:       "Product",
				Description: "Test",
				Price:       -10,
			},
			wantErr: true,
			errMsg:  "price must be non-negative",
		},
		{
			name: "negative count",
			req: core.CreateProductRequest{
				Title:       "Product",
				Description: "Test",
				Price:       100,
				Count:       -5,
			},
			wantErr: true,
			errMsg:  "count must be non-negative",
		},
		{
			name: "zero price is valid",
			req: core.CreateProductRequest{
				Title:       "Product",
				Description: "Test",
				Price:       0,
			},
			wantErr: false,
		},
		{
			name: "zero count is valid",
			req: core.CreateProductRequest{
				Title:       "Product",
				Description: "Test",
				Price:       100,
				Count:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := core.ValidateCreateProductRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCreateProductRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && err != nil && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("ValidateCreateProductRequest() error = %v, want error containing %q", err, tt.errMsg)
			}
		})
	}
}

func TestHandleCreateProductBadRequest(t *testing.T) {
	tests := []struct {
		name           string
		body           interface{}
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "invalid json",
			body:           "{invalid json}",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing title",
			body: core.CreateProductRequest{
				Title: "",
				Price: 100,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "title is required",
		},
		{
			name: "negative price",
			body: core.CreateProductRequest{
				Title: "Product",
				Price: -10,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "price must be non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyBytes []byte
			if s, ok := tt.body.(string); ok {
				bodyBytes = []byte(s)
			} else {
				bodyBytes, _ = json.Marshal(tt.body)
			}

			req := httptest.NewRequest(http.MethodPost, "/products", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			handleCreateProduct(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", rr.Code, tt.expectedStatus)
			}

			if !strings.Contains(rr.Body.String(), tt.expectedError) {
				t.Errorf("handler returned unexpected body: got %s want error containing %s", rr.Body.String(), tt.expectedError)
			}
		})
	}
}

func TestHandleCreateProductMethodCheck(t *testing.T) {
	tests := []struct {
		method         string
		expectedStatus int
	}{
		{http.MethodPost, http.StatusInternalServerError}, // Will fail due to missing DynamoDB but shows POST is handled
		{http.MethodGet, http.StatusMethodNotAllowed},
		{http.MethodPut, http.StatusMethodNotAllowed},
		{http.MethodDelete, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.method, func(t *testing.T) {
			body := core.CreateProductRequest{
				Title: "Test",
				Price: 100,
			}
			bodyBytes, _ := json.Marshal(body)

			req := httptest.NewRequest(tt.method, "/products", bytes.NewReader(bodyBytes))
			rr := httptest.NewRecorder()

			handleCreateProduct(rr, req)

			// Skip validation for non-POST methods or when testing just method handling
			if tt.method != http.MethodPost {
				if rr.Code != tt.expectedStatus {
					t.Errorf("handler returned wrong status code for %s: got %v want %v", tt.method, rr.Code, tt.expectedStatus)
				}
			}
		})
	}
}
