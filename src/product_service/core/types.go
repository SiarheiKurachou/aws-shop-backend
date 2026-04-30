package core

// Product represents a product in the shop.
type Product struct {
	ID          string  `json:"id"`
	Title       string  `json:"title"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// DataFile represents the structure of data.json.
type DataFile struct {
	Products []Product `json:"products"`
}
