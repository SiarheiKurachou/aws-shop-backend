# Create Product API

## Overview

The product service now includes a POST endpoint to create new products with inventory tracking.

## Endpoint

```
POST /products
```

## Request Body

```json
{
  "id": "uuid-string (optional, auto-generated if omitted)",
  "title": "Product Name (required)",
  "description": "Product Description",
  "price": 100,
  "count": 50
}
```

### Fields

| Field | Type | Required | Notes |
|-------|------|----------|-------|
| `id` | string | No | UUID. Auto-generated if omitted |
| `title` | string | Yes | Product name. Cannot be empty or whitespace |
| `description` | string | No | Product description |
| `price` | integer | Yes | Price in cents. Must be ≥ 0 |
| `count` | integer | No | Initial stock count. Defaults to 0. Must be ≥ 0 |

## Response

**Status: 201 Created**

```json
{
  "id": "generated-uuid",
  "title": "Product Name",
  "description": "Product Description",
  "price": 100,
  "count": 50
}
```

## Error Responses

### 400 Bad Request

**Invalid Request Body:**
```json
{
  "error": "Invalid request body: EOF"
}
```

**Missing Title:**
```json
{
  "error": "Failed to create product: title is required"
}
```

**Negative Price:**
```json
{
  "error": "Failed to create product: price must be non-negative"
}
```

**Negative Count:**
```json
{
  "error": "Failed to create product: count must be non-negative"
}
```

### 500 Internal Server Error

**DynamoDB Not Configured:**
```json
{
  "error": "Failed to create product: write operations require DynamoDB to be configured via PRODUCTS_TABLE_NAME and STOCKS_TABLE_NAME environment variables"
}
```

## Validation Rules

1. **Title**: Required, non-empty, cannot be only whitespace
2. **Price**: Must be ≥ 0
3. **Count**: Must be ≥ 0 (defaults to 0 if omitted)

## Transaction Guarantees

Product creation uses DynamoDB transactions to ensure:
- Both product and stock entries are created atomically
- No partial states (product without stock or stock without product)
- All-or-nothing semantics

## Examples

### Create Product with Auto-Generated ID

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Laptop",
    "description": "High-performance laptop",
    "price": 99999,
    "count": 10
  }'
```

Response:
```json
{
  "id": "a1b2c3d4-e5f6-4a5b-9c8d-e7f8g9h0i1j2",
  "title": "Laptop",
  "description": "High-performance laptop",
  "price": 99999,
  "count": 10
}
```

### Create Product with Specific ID

```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "id": "123e4567-e89b-12d3-a456-426614174000",
    "title": "Mouse",
    "description": "Wireless mouse",
    "price": 2999
  }'
```

Response:
```json
{
  "id": "123e4567-e89b-12d3-a456-426614174000",
  "title": "Mouse",
  "description": "Wireless mouse",
  "price": 2999,
  "count": 0
}
```

### Validation Errors

**Missing Title:**
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"price": 100}'
```

Response (400):
```json
{
  "error": "Failed to create product: title is required"
}
```

**Negative Price:**
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Product",
    "price": -50
  }'
```

Response (400):
```json
{
  "error": "Failed to create product: price must be non-negative"
}
```

## Requirements

- **DynamoDB Configuration**: Requires `PRODUCTS_TABLE_NAME` and `STOCKS_TABLE_NAME` environment variables
- **IAM Permissions**: Lambda must have `dynamodb:TransactWriteItems` permission on both tables
- **Local Development**: Use local JSON storage for testing (POST will fail without DynamoDB env vars)

## Integration with CDK

To enable Lambda functions to create products, update [cdk/stack/stack.go](../cdk/stack/stack.go):

```go
// Add write permissions for create operations
productsTable.GrantWriteData(listProductsFn)
stocksTable.GrantWriteData(listProductsFn)
productsTable.GrantWriteData(getProductByIDFn)
stocksTable.GrantWriteData(getProductByIDFn)
```

## Testing

All validation and error cases are covered by tests in [create-product_test.go](../src/product-service/create-product_test.go).

Run tests:
```bash
go test ./src/product-service/... -v
```

## See Also

- [GET /products](./api.md#get-products) - List all products
- [GET /products/{id}](./api.md#get-products-id) - Get product by ID
- [DynamoDB Transactions](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/transaction-apis.html)
