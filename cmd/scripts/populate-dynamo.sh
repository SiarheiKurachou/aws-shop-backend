#!/bin/bash
set -e

# Script to populate DynamoDB tables with test data
# Usage: ./cmd/scripts/populate-dynamo.sh [options]
# Options:
#   --products-table TABLE_NAME   DynamoDB products table name (default: products)
#   --stocks-table TABLE_NAME     DynamoDB stocks table name (default: stocks)
#   --data-file FILE              Path to data.json file (default: cmd/scripts/data.json)
#   --default-stock COUNT         Default stock count per product (default: 100)
#   --clear                       Clear tables before populating
#   --help                        Show this help message

PRODUCTS_TABLE="products"
STOCKS_TABLE="stocks"
DATA_FILE="cmd/scripts/data.json"
DEFAULT_STOCK="100"
CLEAR=false

while [[ $# -gt 0 ]]; do
  case $1 in
    --products-table)
      PRODUCTS_TABLE="$2"
      shift 2
      ;;
    --stocks-table)
      STOCKS_TABLE="$2"
      shift 2
      ;;
    --data-file)
      DATA_FILE="$2"
      shift 2
      ;;
    --default-stock)
      DEFAULT_STOCK="$2"
      shift 2
      ;;
    --clear)
      CLEAR=true
      shift
      ;;
    --help)
      grep '^#' "$0" | tail -n +2 | sed 's/^# *//'
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Build the populate tool
echo "Building populate-dynamo tool..."
go build -o dist/populate-dynamo ./cmd/populate-dynamo

# Run the populate tool
echo "Populating DynamoDB tables..."
CLEAR_FLAG=""
if [ "$CLEAR" = true ]; then
  CLEAR_FLAG="--clear"
fi

./dist/populate-dynamo \
  --products-table "$PRODUCTS_TABLE" \
  --stocks-table "$STOCKS_TABLE" \
  --data-file "$DATA_FILE" \
  --default-stock "$DEFAULT_STOCK" \
  $CLEAR_FLAG

echo "Done!"
