# DynamoDB Population Scripts

This directory contains tools for populating AWS DynamoDB tables with test data for the product service.

## Overview

The population scripts read product data from `src/data.json` and populate two DynamoDB tables:
- **products**: Contains product catalog data (id, title, description, price)
- **stocks**: Contains inventory counts per product (product_id, count)

## Prerequisites

- AWS credentials configured (via environment variables, ~/.aws/credentials, or IAM role)
- DynamoDB tables created (via CDK deployment)
- Go 1.26.2 or later

## Usage

### Using Make (Recommended)

**Populate tables (clear existing data first):**
```bash
make populate-dynamo
```

**Append to tables (keep existing data):**
```bash
make populate-dynamo-append
```

**Just build the tool:**
```bash
make build-populate-dynamo
```

### Using the Go Tool Directly

**Populate with default settings:**
```bash
go run ./cmd/populate-dynamo
```

**Populate with custom table names:**
```bash
go run ./cmd/populate-dynamo \
  --products-table my-products \
  --stocks-table my-stocks
```

**Clear tables before populating:**
```bash
go run ./cmd/populate-dynamo --clear
```

**Populate from custom data file:**
```bash
go run ./cmd/populate-dynamo --data-file path/to/custom-data.json
```

**Set custom default stock count:**
```bash
go run ./cmd/populate-dynamo --default-stock 50
```

### Using the Shell Script

```bash
./scripts/populate-dynamo.sh --clear
./scripts/populate-dynamo.sh --help
```

## Command-Line Options

| Option | Default | Description |
|--------|---------|-------------|
| `--products-table` | `products` | DynamoDB products table name |
| `--stocks-table` | `stocks` | DynamoDB stocks table name |
| `--data-file` | `src/data.json` | Path to source data file |
| `--default-stock` | `100` | Default stock count per product |
| `--clear` | false | Clear tables before populating |

## Data Format

The source data file (default: `src/data.json`) should follow this JSON structure:

```json
{
  "products": [
    {
      "id": "uuid-string",
      "title": "Product Name",
      "description": "Product Description",
      "price": 100
    },
    ...
  ]
}
```

## Environment Variables

### AWS Configuration
- `AWS_REGION`: AWS region (default: from AWS config)
- `AWS_ACCESS_KEY_ID`: AWS access key (if using static credentials)
- `AWS_SECRET_ACCESS_KEY`: AWS secret key (if using static credentials)

## Example Workflows

### Deploy Stack and Populate Tables
```bash
# Deploy infrastructure
make cdk-deploy

# Populate DynamoDB tables
make populate-dynamo
```

### Refresh Test Data
```bash
# Clear and repopulate tables
make populate-dynamo

# Verify by calling API
curl https://your-api-gateway-url/products
```

### Multiple Environments
```bash
# Populate production table
AWS_REGION=us-east-1 go run ./cmd/populate-dynamo \
  --products-table prod-products \
  --stocks-table prod-stocks

# Populate staging table
AWS_REGION=us-west-2 go run ./cmd/populate-dynamo \
  --products-table staging-products \
  --stocks-table staging-stocks
```

## Troubleshooting

### "Table not found" error
- Verify tables exist in DynamoDB
- Check table names match deployment output
- Verify AWS credentials have DynamoDB access

### "Access Denied" error
- Check IAM permissions include `dynamodb:PutItem` and `dynamodb:DeleteItem`
- Verify AWS credentials are configured

### No output after running
- Check AWS region is correct
- Add logging by monitoring CloudWatch

## Testing

To test the populate tool locally with DynamoDB Local:
```bash
# Start DynamoDB Local (requires Docker)
docker run -p 8000:8000 amazon/dynamodb-local

# Run populate tool pointing to local DynamoDB
AWS_ENDPOINT_URL=http://localhost:8000 go run ./cmd/populate-dynamo
```

## Performance

- Inserting 6 sample products + stocks: ~2-5 seconds
- Uses PutItem (non-batched) for simplicity
- For large datasets, consider batch operations

## See Also

- [../src/data.json](../src/data.json) - Sample product data
- [../cdk/stack/](../cdk/stack/) - Infrastructure definitions
- [../src/product-service/](../src/product-service/) - Product service code
