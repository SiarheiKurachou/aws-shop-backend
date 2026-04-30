package main

import "aws-shop-backend/src/product_service/core"

// Type aliases retained for Swagger annotations and existing tests.
type Product = core.Product
type DataFile = core.DataFile

func loadProducts() ([]core.Product, error) { return core.LoadProducts() }
func findProductByID(products []core.Product, id string) (core.Product, bool) {
	return core.FindProductByID(products, id)
}
