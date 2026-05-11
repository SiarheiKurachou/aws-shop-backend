# DynamoDB Population Summary

## What Was Created

I've added comprehensive tooling to populate AWS DynamoDB tables with test data. Here's what was implemented:

### 1. **Go Population Tool** ([cmd/populate-dynamo/main.go](cmd/populate-dynamo/main.go))
- Reads product data from `src/data.json`
- Inserts products into DynamoDB products table
- Creates corresponding stock entries in stocks table
- Supports clearing tables before population (for clean resets)
- Fully typed with AWS SDK v2 DynamoDB client

**Features:**
- Configurable table names
- Configurable default stock counts
- Atomic operations with proper error handling
- Table clearing capability with `--clear` flag

### 2. **Shell Script Wrapper** ([scripts/populate-dynamo.sh](scripts/populate-dynamo.sh))
- User-friendly wrapper around the Go tool
- Automatic build of the tool before running
- Full help system with usage examples
- Support for all command-line options

### 3. **Makefile Targets**
Added three convenient targets to [Makefile](Makefile):
- `make build-populate-dynamo` — Build the tool
- `make populate-dynamo` — Build and run with `--clear` (replace all data)
- `make populate-dynamo-append` — Build and run without clearing (add/update)

### 4. **Documentation** ([docs/POPULATE_DYNAMODB.md](docs/POPULATE_DYNAMODB.md))
- Complete usage guide with examples
- Environment variable configuration
- Troubleshooting section
- Performance notes
- Integration workflows

### 5. **Schema Update**
Fixed [src/product-service/core/types.go](src/product-service/core/types.go):
- Changed `Price` from `float64` to `int` (matches schema requirement)
- Added `Count` field to Product struct (from stocks table join)

## Quick Start

### Option 1: Using Make (Recommended)
```bash
# Clear and repopulate tables
make populate-dynamo

# Add/update without clearing
make populate-dynamo-append
```

### Option 2: Using the shell script
```bash
./scripts/populate-dynamo.sh --clear
```

### Option 3: Direct Go invocation
```bash
go run ./cmd/populate-dynamo --clear
```

## What Gets Populated

From `src/data.json`, the script inserts:
- **6 products** with unique IDs, titles, descriptions, and prices
- **6 stock entries** each with 100 items in stock (configurable via `--default-stock`)

### Example Product Entry:
```json
{
  "id": "7567ec4b-b10c-48c5-9345-fc73c48a80aa",
  "title": "ProductOne",
  "description": "Short Product Description1",
  "price": 24
}
```

### Example Stock Entry:
```
product_id: "7567ec4b-b10c-48c5-9345-fc73c48a80aa"
count: 100
```

## Verification

After populating, test the API endpoints:
```bash
# Get all products (with stock count)
curl https://your-api-gateway-url/products

# Get single product by ID
curl https://your-api-gateway-url/products/7567ec4b-b10c-48c5-9345-fc73c48a80aa
```

Expected response includes the `count` field from stocks table.

## Command-Line Options

| Option | Default | Example |
|--------|---------|---------|
| `--products-table` | `products` | `--products-table prod-products` |
| `--stocks-table` | `stocks` | `--stocks-table prod-stocks` |
| `--data-file` | `src/data.json` | `--data-file custom-data.json` |
| `--default-stock` | `100` | `--default-stock 50` |
| `--clear` | false | `--clear` |

## Environment Variables

Configure AWS credentials and region:
```bash
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=your_key
export AWS_SECRET_ACCESS_KEY=your_secret

make populate-dynamo
```

## Deployment Workflow

Recommended order:
```bash
# 1. Deploy CDK stack (creates tables)
make cdk-deploy

# 2. Populate with test data
make populate-dynamo

# 3. Test API endpoints
curl https://your-api-url/products
```

## Files Modified/Created

- ✅ **Created:** `cmd/populate-dynamo/main.go` — Go tool
- ✅ **Created:** `scripts/populate-dynamo.sh` — Shell wrapper
- ✅ **Created:** `docs/POPULATE_DYNAMODB.md` — Documentation
- ✅ **Updated:** `Makefile` — Added 3 new targets
- ✅ **Fixed:** `src/product-service/core/types.go` — Price type and Count field

## Tests

All existing tests pass:
```bash
go test ./...
# Result: ok      aws-shop-backend/src/product-service (cached)
```

## Next Steps

1. Deploy infrastructure: `make cdk-deploy`
2. Populate DynamoDB: `make populate-dynamo`
3. Test API: `curl https://your-api-url/products`
4. Monitor in CloudWatch
