package handlers

// This package provides HTTP handlers and business logic for the product service.
//
// Handler modules:
// - list_products.go: HandleListProducts, listProductsService
// - get_product_by_id.go: HandleGetProductByID, getProductByIDService
// - create_product.go: HandleCreateProduct, createProductService
//
// Each module contains a thin HTTP handler that delegates to a service function,
// separating HTTP concerns from business logic.
