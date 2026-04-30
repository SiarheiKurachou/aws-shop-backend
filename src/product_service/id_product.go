package main

import (
	"aws-shop-backend/src/product_service/core"
	"net/http"
)

// handleGetProductByID handles the GET /products/{id} request
// @ID get-product-by-id
// @Summary Get a product by ID
// @Description Returns a single product by its ID
// @Tags Products
// @Accept json
// @Produce json
// @Param id path string true "Product ID"
// @Success 200 {object} Product
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /products/{id} [get]
func handleGetProductByID(w http.ResponseWriter, r *http.Request) {
	core.HandleGetProductByID(w, r)
}
