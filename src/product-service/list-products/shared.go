package listproducts

import "aws-shop-backend/src/product-service/common"

// Type aliases retained for Swagger annotations and existing tests.
type Product = common.Product
type DataFile = common.DataFile

func loadProducts() ([]common.Product, error) { return common.LoadProducts() }
func findProductByID(products []common.Product, id string) (common.Product, bool) {
	return common.FindProductByID(products, id)
}
