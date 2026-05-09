package core

import (
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
