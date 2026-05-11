# Create Product Lambda Implementation

## Summary

Added a third Lambda function for the `POST /products` endpoint with transactional product creation, integrated into the CDK stack.

## Changes Made

### 1. Lambda Handler
**File:** [cdk/lambda/create-product/main.go](cdk/lambda/create-product/main.go)

- Lambda handler for `POST /products` endpoint
- Parses JSON request body
- Calls `core.CreateProduct()` with validation
- Returns 201 Created on success
- Returns appropriate error status codes (400, 500)
- Uses build tag `lambda_create_product` to differentiate from local server

**Key Features:**
- Direct call to `core.CreateProduct()` (no httptest wrapper needed)
- Returns JSON with proper headers
- CORS headers included in responses
- Status codes: 201 (created), 400 (validation), 500 (server error)

### 2. CDK Stack Updates
**File:** [cdk/stack/stack.go](cdk/stack/stack.go)

**Added:**
- Third Lambda function: `create-product`
- Write permissions: `GrantWriteData()` on both tables
- API Gateway POST method integration
- CloudFormation output for Lambda name
- POST added to CORS AllowMethods

**Permissions:**
- `productsTable.GrantWriteData(createProductFn)`
- `stocksTable.GrantWriteData(createProductFn)`

**API Gateway Changes:**
```
GET  /products      → ListProducts Lambda (existing)
POST /products      → CreateProduct Lambda (new)
GET  /products/{id} → GetProductByID Lambda (existing)
```

**CORS Update:**
```
Allowed Methods: GET, POST, OPTIONS (updated from GET, OPTIONS)
```

### 3. Build Configuration
**File:** [Makefile](Makefile)

Updated `build-lambdas` target:
```bash
# Added create-product build directory
mkdir -p cdk/dist/create-product

# Added create-product Lambda build with build tag
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build \
  -buildvcs=false \
  -tags lambda_create_product \
  -o cdk/dist/create-product/bootstrap \
  ./cdk/lambda/create-product
```

## Deployment

### Build
```bash
make build-lambdas
```

Creates three Lambda bootstrap binaries:
- `cdk/dist/get-products-list/bootstrap` (20.7 MB)
- `cdk/dist/get-product-by-id/bootstrap` (17.5 MB)
- `cdk/dist/create-product/bootstrap` (17.5 MB)

### Deploy
```bash
make cdk-deploy
```

Deploys infrastructure with:
- All three Lambda functions
- Read-only access for GET Lambdas
- Write access for POST Lambda
- API Gateway with GET/POST routes
- DynamoDB tables (products, stocks)

## API Usage

### Create Product
```bash
curl -X POST https://your-api-gateway-url/products \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Laptop",
    "description": "Gaming laptop",
    "price": 99999,
    "count": 5
  }'
```

### Response (201 Created)
```json
{
  "id": "a1b2c3d4-e5f6-4a5b-9c8d-e7f8g9h0i1j2",
  "title": "Laptop",
  "description": "Gaming laptop",
  "price": 99999,
  "count": 5
}
```

## Testing

All tests pass:
```bash
go test ./src/product-service/...
# ✓ 16 tests pass
```

## Stack Architecture

```
API Gateway
├── GET  /products      → get-products-list Lambda (ReadOnly)
├── POST /products      → create-product Lambda (WriteOnly)
└── GET  /products/{id} → get-product-by-id Lambda (ReadOnly)

DynamoDB
├── products table (read+write)
│   └── Partition Key: id (STRING)
└── stocks table (read+write)
    └── Partition Key: product_id (STRING)
```

## Environment Variables

Lambda functions receive:
- `PRODUCTS_TABLE_NAME` = "products"
- `STOCKS_TABLE_NAME` = "stocks"

## Security

**IAM Permissions:**
- GET Lambdas: `dynamodb:GetItem`, `dynamodb:Scan`
- POST Lambda: `dynamodb:PutItem`, `dynamodb:TransactWriteItems`

## Files Modified/Created

| File | Action |
|------|--------|
| `cdk/lambda/create-product/main.go` | Created |
| `cdk/stack/stack.go` | Modified |
| `Makefile` | Modified |

## Next Steps

1. Deploy stack: `make cdk-deploy`
2. Test endpoints via API Gateway
3. Monitor CloudWatch logs
4. Verify DynamoDB entries in both tables

## See Also

- [docs/CREATE_PRODUCT.md](../docs/CREATE_PRODUCT.md) — Full API documentation
- [docs/CREATE_PRODUCT_IMPLEMENTATION.md](../docs/CREATE_PRODUCT_IMPLEMENTATION.md) — Core implementation details
- [docs/POPULATE_DYNAMODB.md](../docs/POPULATE_DYNAMODB.md) — Initial data population
