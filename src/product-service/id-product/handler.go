package idproduct

import (
	corehandlers "aws-shop-backend/src/product-service/core/handlers"
	"net/http"
)

// HandleGetProductByID handles the GET /products/{id} request
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
func HandleGetProductByID(w http.ResponseWriter, r *http.Request) {
	corehandlers.HandleGetProductByID(w, r)
}
