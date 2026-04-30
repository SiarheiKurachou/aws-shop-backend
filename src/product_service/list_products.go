package main

import (
	"aws-shop-backend/src/product_service/core"
	"net/http"
)

// handleListProducts handles the GET /products request
// @ID get-products-list
// @Summary Get list of all products
// @Description Returns a list of all available products
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {array} Product
// @Failure 500 {object} map[string]interface{}
// @Router /products [get]
func handleListProducts(w http.ResponseWriter, r *http.Request) {
	core.HandleListProducts(w, r)
}
