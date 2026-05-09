//go:build !lambda_list_products && !lambda_get_product_by_id

package main

import (
	"aws-shop-backend/src/product-service/core"
	_ "aws-shop-backend/src/product-service/docs"
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
			handleCreateProduct(w, r)
		} else if r.Method == "GET" {
			handleListProducts(w, r)
		} else {
			core.WriteError(w, http.StatusMethodNotAllowed, "Method not allowed")
		}
	})

	mux.HandleFunc("/products/", handleGetProductByID)
	mux.Handle("/swagger/", httpSwagger.WrapHandler)

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
