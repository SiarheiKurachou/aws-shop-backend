package listproducts

import "net/http"

// HandleListProducts handles the GET /products request
// @ID get-products-list
// @Summary Get list of all products
// @Description Returns a list of all available products
// @Tags Products
// @Accept json
// @Produce json
// @Success 200 {array} Product
// @Failure 500 {object} map[string]interface{}
// @Router /products [get]
func HandleListProducts(w http.ResponseWriter, r *http.Request) {
	handleListProducts(w, r)
}
