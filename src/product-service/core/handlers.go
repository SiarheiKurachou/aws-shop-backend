package core

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleListProducts is the canonical HTTP handler for GET /products.
func HandleListProducts(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GET /products request")

	products, err := LoadProducts()
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to load products")
		return
	}

	WriteJSON(w, http.StatusOK, products)
}

// HandleGetProductByID is the canonical HTTP handler for GET /products/{id}.
func HandleGetProductByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling GET /products/{id} request")

	productID := ProductIDFromRequest(r)
	if productID == "" {
		WriteError(w, http.StatusBadRequest, "Product id is required")
		return
	}

	product, found, err := LoadProductByID(productID)
	if err != nil {
		WriteError(w, http.StatusInternalServerError, "Failed to load products")
		return
	}

	if !found {
		WriteError(w, http.StatusNotFound, "Product not found")
		return
	}

	WriteJSON(w, http.StatusOK, product)
}

// HandleCreateProduct is the canonical HTTP handler for POST /products.
func HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	log.Printf("Handling POST /products request")

	if r.Method != http.MethodPost {
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	// Create product
	product, err := CreateProduct(req)
	if err != nil {
		// Check if it's a validation error or system error
		log.Printf("Error creating product: %v", err)

		// Determine HTTP status code
		statusCode := http.StatusInternalServerError
		if err.Error() == "title is required" ||
			err.Error() == "price must be non-negative" ||
			err.Error() == "count must be non-negative" {
			statusCode = http.StatusBadRequest
		}

		WriteError(w, statusCode, "Failed to create product: "+err.Error())
		return
	}

	WriteJSON(w, http.StatusCreated, CreateProductResponse{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Count:       product.Count,
	})
}
