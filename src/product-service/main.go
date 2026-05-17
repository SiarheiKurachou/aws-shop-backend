//go:build !lambda_list_products && !lambda_get_product_by_id

package main

import (
	"aws-shop-backend/src/httphelpers"
	createproduct "aws-shop-backend/src/product-service/create-product"
	idproduct "aws-shop-backend/src/product-service/id-product"
	listproducts "aws-shop-backend/src/product-service/list-products"
	"log"
	"net/http"

	httpSwagger "github.com/swaggo/http-swagger"
)

// @title AWS Shop Backend API
// @version 1.0
// @description Product service API
// @BasePath /

func main() {
	mux := http.NewServeMux()

	// Handle /products with method-based routing
	mux.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			createproduct.HandleCreateProduct(w, r)
		} else if r.Method == "GET" {
			listproducts.HandleListProducts(w, r)
		} else {
			httphelpers.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/products/", idproduct.HandleGetProductByID)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
