package main

import (
	"aws-shop-backend/src/product-service/core"
	"net/http"
)

// handleCreateProduct handles the POST /products request
// @ID create-product
// @Summary Create a new product
// @Description Creates a new product with initial stock count
// @Tags Products
// @Accept json
// @Produce json
// @Param body body core.CreateProductRequest true "Product data"
// @Success 201 {object} core.CreateProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /products [post]
func handleCreateProduct(w http.ResponseWriter, r *http.Request) {
	core.HandleCreateProduct(w, r)
}
