package core

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamodbTypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

var (
	dynamoClientOnce sync.Once
	dynamoClient     *dynamodb.Client
	dynamoClientErr  error
)

type dbProduct struct {
	ID          string `dynamodbav:"id"`
	Title       string `dynamodbav:"title"`
	Description string `dynamodbav:"description"`
	Price       int    `dynamodbav:"price"`
}

type dbStock struct {
	ProductID string `dynamodbav:"product_id"`
	Count     int    `dynamodbav:"count"`
}

func ProductIDFromRequest(r *http.Request) string {
	productID := strings.TrimPrefix(r.URL.Path, "/products/")
	if productID == r.URL.Path || productID == "" {
		return r.URL.Query().Get("id")
	}

	return productID
}

func FindProductByID(products []Product, id string) (Product, bool) {
	for _, product := range products {
		if product.ID == id {
			return product, true
		}
	}

	return Product{}, false
}

func tableNamesFromEnv() (string, string, bool, error) {
	productsTable := os.Getenv("PRODUCTS_TABLE_NAME")
	stocksTable := os.Getenv("STOCKS_TABLE_NAME")

	if productsTable == "" && stocksTable == "" {
		return "", "", false, nil
	}

	if productsTable == "" || stocksTable == "" {
		return "", "", false, fmt.Errorf("both PRODUCTS_TABLE_NAME and STOCKS_TABLE_NAME must be set")
	}

	return productsTable, stocksTable, true, nil
}

func getDynamoClient(ctx context.Context) (*dynamodb.Client, error) {
	dynamoClientOnce.Do(func() {
		cfg, err := config.LoadDefaultConfig(ctx)
		if err != nil {
			dynamoClientErr = fmt.Errorf("load aws config: %w", err)
			return
		}

		dynamoClient = dynamodb.NewFromConfig(cfg)
	})

	if dynamoClientErr != nil {
		return nil, dynamoClientErr
	}

	return dynamoClient, nil
}

func readAllProductsFromDynamo(ctx context.Context, client *dynamodb.Client, tableName string) ([]dbProduct, error) {
	paginator := dynamodb.NewScanPaginator(client, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	products := make([]dbProduct, 0)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scan products table: %w", err)
		}

		var pageProducts []dbProduct
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &pageProducts); err != nil {
			return nil, fmt.Errorf("unmarshal products table items: %w", err)
		}

		products = append(products, pageProducts...)
	}

	return products, nil
}

func readAllStocksFromDynamo(ctx context.Context, client *dynamodb.Client, tableName string) (map[string]int, error) {
	paginator := dynamodb.NewScanPaginator(client, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})

	stocksByProductID := make(map[string]int)
	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("scan stocks table: %w", err)
		}

		var pageStocks []dbStock
		if err := attributevalue.UnmarshalListOfMaps(page.Items, &pageStocks); err != nil {
			return nil, fmt.Errorf("unmarshal stocks table items: %w", err)
		}

		for _, stock := range pageStocks {
			stocksByProductID[stock.ProductID] = stock.Count
		}
	}

	return stocksByProductID, nil
}

func loadProductsFromDynamo(ctx context.Context, productsTable string, stocksTable string) ([]Product, error) {
	client, err := getDynamoClient(ctx)
	if err != nil {
		return nil, err
	}

	products, err := readAllProductsFromDynamo(ctx, client, productsTable)
	if err != nil {
		return nil, err
	}

	stocksByProductID, err := readAllStocksFromDynamo(ctx, client, stocksTable)
	if err != nil {
		return nil, err
	}

	result := make([]Product, 0, len(products))
	for _, p := range products {
		result = append(result, Product{
			ID:          p.ID,
			Title:       p.Title,
			Description: p.Description,
			Price:       p.Price,
			Count:       stocksByProductID[p.ID],
		})
	}

	return result, nil
}

func loadProductByIDFromDynamo(ctx context.Context, productsTable string, stocksTable string, productID string) (Product, bool, error) {
	client, err := getDynamoClient(ctx)
	if err != nil {
		return Product{}, false, err
	}

	productResp, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(productsTable),
		Key: map[string]dynamodbTypes.AttributeValue{
			"id": &dynamodbTypes.AttributeValueMemberS{Value: productID},
		},
	})
	if err != nil {
		return Product{}, false, fmt.Errorf("get product by id: %w", err)
	}

	if len(productResp.Item) == 0 {
		return Product{}, false, nil
	}

	var dbp dbProduct
	if err := attributevalue.UnmarshalMap(productResp.Item, &dbp); err != nil {
		return Product{}, false, fmt.Errorf("unmarshal product item: %w", err)
	}

	count := 0
	stockResp, err := client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(stocksTable),
		Key: map[string]dynamodbTypes.AttributeValue{
			"product_id": &dynamodbTypes.AttributeValueMemberS{Value: productID},
		},
	})
	if err != nil {
		return Product{}, false, fmt.Errorf("get stock by product id: %w", err)
	}

	if len(stockResp.Item) > 0 {
		var dbs dbStock
		if err := attributevalue.UnmarshalMap(stockResp.Item, &dbs); err != nil {
			return Product{}, false, fmt.Errorf("unmarshal stock item: %w", err)
		}
		count = dbs.Count
	}

	return Product{
		ID:          dbp.ID,
		Title:       dbp.Title,
		Description: dbp.Description,
		Price:       dbp.Price,
		Count:       count,
	}, true, nil
}

func loadProductsFromJSON() ([]Product, error) {
	var dataPath string

	if envPath := os.Getenv("PRODUCTS_DATA_FILE"); envPath != "" {
		if _, err := os.Stat(envPath); err == nil {
			dataPath = envPath
		} else {
			return nil, fmt.Errorf("PRODUCTS_DATA_FILE path does not exist: %s", envPath)
		}
	} else {
		var candidates []string
		candidates = append(candidates, "cmd/scripts/data.json", "src/data.json", "../data.json", "data.json")

		execPath := filepath.Join(filepath.Dir(os.Args[0]), "../data.json")
		candidates = append(candidates, execPath)

		if _, thisFile, _, ok := runtime.Caller(0); ok {
			candidates = append(candidates, filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "cmd", "scripts", "data.json"))
			candidates = append(candidates, filepath.Join(filepath.Dir(thisFile), "..", "..", "data.json"))
		}

		for _, p := range candidates {
			if _, err := os.Stat(p); err == nil {
				dataPath = p
				break
			}
		}

		if dataPath == "" {
			return nil, fmt.Errorf("data.json not found in expected locations")
		}
	}

	data, err := os.ReadFile(dataPath)
	if err != nil {
		log.Printf("Error reading data.json: %v", err)
		return nil, err
	}

	var dataFile DataFile
	if err := json.Unmarshal(data, &dataFile); err != nil {
		log.Printf("Error unmarshaling data.json: %v", err)
		return nil, err
	}

	return dataFile.Products, nil
}

// LoadProducts loads product data from data.json.
// It first checks the PRODUCTS_DATA_FILE environment variable.
// If not set, it searches in standard locations.
func LoadProducts() ([]Product, error) {
	productsTable, stocksTable, configured, err := tableNamesFromEnv()
	if err != nil {
		return nil, err
	}

	if configured {
		return loadProductsFromDynamo(context.Background(), productsTable, stocksTable)
	}

	return loadProductsFromJSON()
}

// LoadProductByID resolves a product by id from DynamoDB when configured.
// In local mode, it falls back to JSON-backed lookup.
func LoadProductByID(productID string) (Product, bool, error) {
	productsTable, stocksTable, configured, err := tableNamesFromEnv()
	if err != nil {
		return Product{}, false, err
	}

	if configured {
		return loadProductByIDFromDynamo(context.Background(), productsTable, stocksTable, productID)
	}

	products, err := loadProductsFromJSON()
	if err != nil {
		return Product{}, false, err
	}

	product, found := FindProductByID(products, productID)
	return product, found, nil
}

// ValidateCreateProductRequest validates a create product request.
func ValidateCreateProductRequest(req CreateProductRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if req.Price < 0 {
		return fmt.Errorf("price must be non-negative")
	}

	if req.Count < 0 {
		return fmt.Errorf("count must be non-negative")
	}

	return nil
}

// CreateProductInDynamo creates a new product and stock entry in DynamoDB using a transaction.
func CreateProductInDynamo(ctx context.Context, productsTable string, stocksTable string, req CreateProductRequest) (Product, error) {
	client, err := getDynamoClient(ctx)
	if err != nil {
		return Product{}, err
	}

	// Generate UUID if not provided
	productID := req.ID
	if productID == "" {
		productID = uuid.New().String()
	}

	// Default stock count to 0 if not provided
	count := req.Count
	if count < 0 {
		count = 0
	}

	// Prepare product item
	productItem := dbProduct{
		ID:          productID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
	}

	productAV, err := attributevalue.MarshalMap(productItem)
	if err != nil {
		return Product{}, fmt.Errorf("marshal product: %w", err)
	}

	// Prepare stock item
	stockItem := dbStock{
		ProductID: productID,
		Count:     count,
	}

	stockAV, err := attributevalue.MarshalMap(stockItem)
	if err != nil {
		return Product{}, fmt.Errorf("marshal stock: %w", err)
	}

	// Use transact write to insert both items atomically
	_, err = client.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []dynamodbTypes.TransactWriteItem{
			{
				Put: &dynamodbTypes.Put{
					TableName: aws.String(productsTable),
					Item:      productAV,
				},
			},
			{
				Put: &dynamodbTypes.Put{
					TableName: aws.String(stocksTable),
					Item:      stockAV,
				},
			},
		},
	})

	if err != nil {
		return Product{}, fmt.Errorf("transact write items: %w", err)
	}

	return Product{
		ID:          productID,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Count:       count,
	}, nil
}

// CreateProduct creates a new product with validation.
// Uses DynamoDB if configured, otherwise returns an error (JSON doesn't support writes).
func CreateProduct(req CreateProductRequest) (Product, error) {
	// Validate request
	if err := ValidateCreateProductRequest(req); err != nil {
		return Product{}, err
	}

	productsTable, stocksTable, configured, err := tableNamesFromEnv()
	if err != nil {
		return Product{}, err
	}

	if !configured {
		return Product{}, fmt.Errorf("write operations require DynamoDB to be configured via PRODUCTS_TABLE_NAME and STOCKS_TABLE_NAME environment variables")
	}

	return CreateProductInDynamo(context.Background(), productsTable, stocksTable, req)
}
