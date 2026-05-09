package core

import (
	"encoding/json"
	"log"
	"net/http"
)

// HandleListProducts is the canonical HTTP handler for GET /products.
func HandleListProducts(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ListProducts] START - Method=%s, Path=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)

	products, err := LoadProducts()
	if err != nil {
		log.Printf("[ListProducts] ERROR - Failed to load products: %v", err)
		WriteError(w, http.StatusInternalServerError, "Failed to load products")
		return
	}

	log.Printf("[ListProducts] SUCCESS - Loaded %d products, responding with status=%d", len(products), http.StatusOK)
	WriteJSON(w, http.StatusOK, products)
}

// HandleGetProductByID is the canonical HTTP handler for GET /products/{id}.
func HandleGetProductByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetProductByID] START - Method=%s, Path=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)

	productID := ProductIDFromRequest(r)
	log.Printf("[GetProductByID] EXTRACT - ProductID=%s", productID)

	if productID == "" {
		log.Printf("[GetProductByID] VALIDATION_ERROR - Product ID is empty")
		WriteError(w, http.StatusBadRequest, "Product id is required")
		return
	}

	product, found, err := LoadProductByID(productID)
	if err != nil {
		log.Printf("[GetProductByID] ERROR - Failed to load product %s: %v", productID, err)
		WriteError(w, http.StatusInternalServerError, "Failed to load products")
		return
	}

	if !found {
		log.Printf("[GetProductByID] NOT_FOUND - Product %s does not exist", productID)
		WriteError(w, http.StatusNotFound, "Product not found")
		return
	}

	log.Printf("[GetProductByID] SUCCESS - Found product %s, Title=%s, Price=%d, responding with status=%d", productID, product.Title, product.Price, http.StatusOK)
	WriteJSON(w, http.StatusOK, product)
}

// HandleCreateProduct is the canonical HTTP handler for POST /products.
func HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CreateProduct] START - Method=%s, Path=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Printf("[CreateProduct] VALIDATION_ERROR - Invalid HTTP method: %s (expected POST)", r.Method)
		WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse request body
	var req CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CreateProduct] PARSE_ERROR - Failed to decode request body: %v", err)
		WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	log.Printf("[CreateProduct] PARSED - Title=%s, Price=%d, Count=%d", req.Title, req.Price, req.Count)

	// Create product
	product, err := CreateProduct(req)
	if err != nil {
		log.Printf("[CreateProduct] CREATION_ERROR - Failed to create product with Title=%s: %v", req.Title, err)

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

	log.Printf("[CreateProduct] SUCCESS - Created product ID=%s, Title=%s, Price=%d, responding with status=%d", product.ID, product.Title, product.Price, http.StatusCreated)
	WriteJSON(w, http.StatusCreated, CreateProductResponse{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Count:       product.Count,
	})
}
