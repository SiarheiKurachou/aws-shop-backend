package createproduct

import (
	corehandlers "aws-shop-backend/src/product-service/core/handlers"
	"net/http"
)

// HandleCreateProduct handles the POST /products request
// @ID create-product
// @Summary Create a new product
// @Description Creates a new product with initial stock count
// @Tags Products
// @Accept json
// @Produce json
// @Param body body CreateProductRequest true "Product data"
// @Success 201 {object} CreateProductResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /products [post]
func HandleCreateProduct(w http.ResponseWriter, r *http.Request) {
	corehandlers.HandleCreateProduct(w, r)
}
