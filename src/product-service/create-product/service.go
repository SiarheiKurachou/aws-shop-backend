package createproduct

import (
	"aws-shop-backend/src/httphelpers"
	"aws-shop-backend/src/product-service/common"
	"encoding/json"
	"log"
	"net/http"
)

// createProductService creates a new product from the given request.
func createProductService(req common.CreateProductRequest) (common.CreateProductResponse, error) {
	log.Printf("[createProductService] Creating product - Title=%s, Price=%d, Count=%d", req.Title, req.Price, req.Count)

	product, err := common.CreateProduct(req)
	if err != nil {
		return common.CreateProductResponse{}, err
	}

	return common.CreateProductResponse{
		ID:          product.ID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Count:       product.Count,
	}, nil
}

func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	log.Printf("[CreateProduct] START - Method=%s, Path=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)

	if r.Method != http.MethodPost {
		log.Printf("[CreateProduct] VALIDATION_ERROR - Invalid HTTP method: %s (expected POST)", r.Method)
		httphelpers.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req common.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("[CreateProduct] PARSE_ERROR - %v", err)
		httphelpers.WriteError(w, http.StatusBadRequest, "Invalid request body: "+err.Error())
		return
	}

	log.Printf("[CreateProduct] PARSED - Title=%s, Price=%d, Count=%d", req.Title, req.Price, req.Count)

	product, err := createProductService(req)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "title is required" ||
			err.Error() == "price must be non-negative" ||
			err.Error() == "count must be non-negative" {
			statusCode = http.StatusBadRequest
		}

		log.Printf("[CreateProduct] ERROR (%d) - %v", statusCode, err)
		httphelpers.WriteError(w, statusCode, "Failed to create product: "+err.Error())
		return
	}

	log.Printf("[CreateProduct] SUCCESS - Created product ID=%s, Title=%s, Price=%d", product.ID, product.Title, product.Price)
	httphelpers.WriteJSON(w, http.StatusCreated, product)
}
