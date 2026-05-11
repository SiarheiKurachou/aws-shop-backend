# Create Product POST Handler Implementation

## Overview

Added a complete POST `/products` endpoint with request validation, automatic UUID generation, and transactional writes to DynamoDB. Both product and stock entries are created atomically.

## Changes Made

### 1. **Type Definitions** ([src/product-service/core/types.go](src/product-service/core/types.go))
- Added `CreateProductRequest` struct with optional ID and stock count
- Added `CreateProductResponse` struct for API responses

### 2. **Core Helper Functions** ([src/product-service/core/helpers.go](src/product-service/core/helpers.go))
- `ValidateCreateProductRequest()` — Validates title, price, and stock count
- `CreateProductInDynamo()` — Uses DynamoDB transactions (`TransactWriteItems`) to atomically insert product and stock
- `CreateProduct()` — Public API that handles validation and delegates to DynamoDB or returns error for JSON-only mode

**Key Features:**
- Automatic UUID generation (v4) if not provided
- Transactional insert ensures product and stock are created together
- Validates: non-empty title, non-negative price/count
- Defaults stock count to 0 if omitted

### 3. **Handler Function** ([src/product-service/core/handlers.go](src/product-service/core/handlers.go))
- `HandleCreateProduct()` — HTTP handler that:
  - Parses JSON request body
  - Calls `CreateProduct()` with validation
  - Returns 201 Created on success
  - Returns 400 Bad Request for validation errors
  - Returns 500 for system errors

### 4. **Route Wrapper** ([src/product-service/create-product.go](src/product-service/create-product.go))
- `handleCreateProduct()` — Wrapper matching the handler naming convention
- Includes Swagger/OpenAPI annotations for `/products` POST endpoint

### 5. **Tests** ([src/product-service/create-product_test.go](src/product-service/create-product_test.go))
- `TestValidateCreateProductRequest()` — 8 test cases covering all validation rules
- `TestHandleCreateProductBadRequest()` — 3 test cases for error responses
- `TestHandleCreateProductMethodCheck()` — Verifies method handling

### 6. **Route Registration** ([src/product-service/main.go](src/product-service/main.go))
- Updated `/products` route to handle both GET and POST
- GET → `handleListProducts()`
- POST → `handleCreateProduct()`

### 7. **Dependencies**
- Added `github.com/google/uuid` for UUID generation

### 8. **Documentation**
- Created [docs/CREATE_PRODUCT.md](docs/CREATE_PRODUCT.md) with:
  - Endpoint specification
  - Request/response examples
  - Validation rules
  - Error codes
  - curl examples
  - Integration requirements

## Usage

### Local Development (Without DynamoDB)
Tests work without DynamoDB. Production operations require DynamoDB.

### Production (With DynamoDB)

**Set Environment Variables:**
```bash
export PRODUCTS_TABLE_NAME=products
export STOCKS_TABLE_NAME=stocks
```

**Create Product:**
```bash
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Laptop",
    "description": "Gaming laptop",
    "price": 99999,
    "count": 5
  }'
```

**Response (201 Created):**
```json
{
  "id": "a1b2c3d4-e5f6-4a5b-9c8d-e7f8g9h0i1j2",
  "title": "Laptop",
  "description": "Gaming laptop",
  "price": 99999,
  "count": 5
}
```

## Validation

All requests are validated:

| Rule | Validation |
|------|-----------|
| Title | Required, non-empty, no whitespace-only |
| Price | Required, must be ≥ 0 |
| Count | Optional, defaults to 0, must be ≥ 0 |

## Transaction Guarantees

Uses DynamoDB `TransactWriteItems` to ensure atomicity:
- ✓ Both product and stock inserted together
- ✓ All-or-nothing semantics
- ✓ No partial states
- ✓ Automatic rollback on error

## CDK Configuration Required

To enable Lambda functions to create products, update [cdk/stack/stack.go](cdk/stack/stack.go):

```go
// Add write permissions for create operations
productsTable.GrantWriteData(listProductsFn)
stocksTable.GrantWriteData(listProductsFn)
productsTable.GrantWriteData(getProductByIDFn)
stocksTable.GrantWriteData(getProductByIDFn)
```

Then redeploy:
```bash
make cdk-deploy
```

## Testing

All tests pass with comprehensive coverage:

```bash
go test ./src/product-service/... -v
```

**Test Coverage:**
- ✓ Valid product creation
- ✓ Missing/empty title validation
- ✓ Negative price validation
- ✓ Negative count validation
- ✓ Invalid JSON handling
- ✓ Method validation (POST vs GET/PUT/DELETE)
- ✓ Auto-generated UUID
- ✓ Default stock count (0)

## Error Responses

| Status | Scenario | Example |
|--------|----------|---------|
| 201 | Success | Product created |
| 400 | Bad Request | Invalid JSON, missing title, negative price |
| 500 | Server Error | DynamoDB misconfigured |

## HTTP Methods

| Method | Path | Handler |
|--------|------|---------|
| GET | `/products` | `HandleListProducts()` |
| POST | `/products` | `HandleCreateProduct()` |
| GET | `/products/{id}` | `HandleGetProductByID()` |

## Integration Workflow

```bash
# 1. Deploy infrastructure
make cdk-deploy

# 2. Populate initial data
make populate-dynamo

# 3. Create new products via API
curl -X POST https://your-api-url/products \
  -H "Content-Type: application/json" \
  -d '{"title":"New Product","price":5000}'

# 4. Verify
curl https://your-api-url/products
```

## Files Modified/Created

| File | Action | Purpose |
|------|--------|---------|
| `src/product-service/core/types.go` | Modified | Added request/response types |
| `src/product-service/core/helpers.go` | Modified | Added validation and transaction logic |
| `src/product-service/core/handlers.go` | Modified | Added POST handler |
| `src/product-service/create-product.go` | Created | Route wrapper |
| `src/product-service/create-product_test.go` | Created | Test suite |
| `src/product-service/main.go` | Modified | Updated route registration |
| `docs/CREATE_PRODUCT.md` | Created | API documentation |
| `go.mod` | Modified | Added google/uuid |

## Next Steps

1. ✅ Implement POST handler with validation
2. ✅ Add tests
3. ⏳ Update CDK to grant write permissions
4. ⏳ Deploy stack: `make cdk-deploy`
5. ⏳ Test in production

## See Also

- [docs/CREATE_PRODUCT.md](docs/CREATE_PRODUCT.md) — Full API documentation
- [docs/POPULATE_DYNAMODB.md](docs/POPULATE_DYNAMODB.md) — Populate initial data
- [DynamoDB Transactions](https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/transaction-apis.html)
- [Google UUID](https://github.com/google/uuid)
