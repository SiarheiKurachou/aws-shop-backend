package handlers

import (
	"aws-shop-backend/src/httphelpers"
	"aws-shop-backend/src/product-service/core"
	"errors"
	"log"
	"net/http"
)

// getProductByIDService retrieves a single product by ID.
// Returns an error if the ID is invalid or the product is not found.
func getProductByIDService(productID string) (core.Product, error) {
	if productID == "" {
		return core.Product{}, errors.New("Product id is required")
	}

	product, found, err := core.LoadProductByID(productID)
	if err != nil {
		log.Printf("[getProductByIDService] Failed to load product %s: %v", productID, err)
		return core.Product{}, errors.New("Failed to load products")
	}

	if !found {
		return core.Product{}, errors.New("Product not found")
	}

	return product, nil
}

// HandleGetProductByID is the canonical HTTP handler for GET /products/{id}.
func HandleGetProductByID(w http.ResponseWriter, r *http.Request) {
	log.Printf("[GetProductByID] START - Method=%s, Path=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)

	productID := core.ProductIDFromRequest(r)
	log.Printf("[GetProductByID] EXTRACT - ProductID=%s", productID)

	product, err := getProductByIDService(productID)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "Product id is required" {
			statusCode = http.StatusBadRequest
		} else if err.Error() == "Product not found" {
			statusCode = http.StatusNotFound
		}

		log.Printf("[GetProductByID] ERROR (%d) - %v", statusCode, err)
		httphelpers.WriteError(w, statusCode, err.Error())
		return
	}

	log.Printf("[GetProductByID] SUCCESS - Found product %s, Title=%s, Price=%d", productID, product.Title, product.Price)
	httphelpers.WriteJSON(w, http.StatusOK, product)
}
