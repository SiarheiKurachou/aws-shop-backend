package common

// Product represents a product in the shop.
type Product struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Count       int    `json:"count"`
}

// DataFile represents the structure of data.json.
type DataFile struct {
	Products []Product `json:"products"`
}

// CreateProductRequest represents a request to create a new product.
type CreateProductRequest struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Count       int    `json:"count,omitempty"`
}

// CreateProductResponse represents the response after creating a product.
type CreateProductResponse struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int    `json:"price"`
	Count       int    `json:"count"`
}
