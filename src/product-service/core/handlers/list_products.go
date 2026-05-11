package handlers

import (
	"aws-shop-backend/src/httphelpers"
	"aws-shop-backend/src/product-service/core"
	"log"
	"net/http"
)

// listProductsService retrieves all products from the data source.
func listProductsService() ([]core.Product, error) {
	log.Printf("[listProductsService] Loading all products")
	return core.LoadProducts()
}

// HandleListProducts is the canonical HTTP handler for GET /products.
func HandleListProducts(w http.ResponseWriter, r *http.Request) {
	log.Printf("[ListProducts] START - Method=%s, Path=%s, RemoteAddr=%s", r.Method, r.URL.Path, r.RemoteAddr)

	products, err := listProductsService()
	if err != nil {
		log.Printf("[ListProducts] ERROR - %v", err)
		httphelpers.WriteError(w, http.StatusInternalServerError, "Failed to load products")
		return
	}

	log.Printf("[ListProducts] SUCCESS - Loaded %d products", len(products))
	httphelpers.WriteJSON(w, http.StatusOK, products)
}
