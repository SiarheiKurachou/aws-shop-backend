package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Product struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Price       int    `json:"price"`
}

type DataFile struct {
	Products []Product `json:"products"`
}

type DBProduct struct {
	ID          string `dynamodbav:"id"`
	Title       string `dynamodbav:"title"`
	Description string `dynamodbav:"description"`
	Price       int    `dynamodbav:"price"`
}

type DBStock struct {
	ProductID string `dynamodbav:"product_id"`
	Count     int    `dynamodbav:"count"`
}

func main() {
	productsTable := flag.String("products-table", "products", "DynamoDB products table name")
	stocksTable := flag.String("stocks-table", "stocks", "DynamoDB stocks table name")
	dataFilePath := flag.String("data-file", "cmd/scripts/data.json", "Path to data.json file")
	defaultStock := flag.Int("default-stock", 100, "Default stock count per product")
	clearTables := flag.Bool("clear", false, "Clear tables before populating")
	flag.Parse()

	ctx := context.Background()

	// Load AWS config
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatalf("Unable to load SDK config: %v", err)
	}

	client := dynamodb.NewFromConfig(cfg)

	// Read data file
	data, err := os.ReadFile(*dataFilePath)
	if err != nil {
		log.Fatalf("Failed to read data file %s: %v", *dataFilePath, err)
	}

	var dataFile DataFile
	if err := json.Unmarshal(data, &dataFile); err != nil {
		log.Fatalf("Failed to parse data file: %v", err)
	}

	log.Printf("Loaded %d products from %s", len(dataFile.Products), *dataFilePath)

	// Clear tables if requested
	if *clearTables {
		if err := clearTable(ctx, client, *productsTable); err != nil {
			log.Fatalf("Failed to clear products table: %v", err)
		}
		log.Printf("Cleared %s", *productsTable)

		if err := clearTable(ctx, client, *stocksTable); err != nil {
			log.Fatalf("Failed to clear stocks table: %v", err)
		}
		log.Printf("Cleared %s", *stocksTable)
	}

	// Insert products
	for _, product := range dataFile.Products {
		dbProduct := DBProduct{
			ID:          product.ID,
			Title:       product.Title,
			Description: product.Description,
			Price:       product.Price,
		}

		av, err := attributevalue.MarshalMap(dbProduct)
		if err != nil {
			log.Fatalf("Failed to marshal product %s: %v", product.ID, err)
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(*productsTable),
			Item:      av,
		})
		if err != nil {
			log.Fatalf("Failed to insert product %s: %v", product.ID, err)
		}

		log.Printf("Inserted product: %s (%s)", product.Title, product.ID)
	}

	// Insert stocks
	for _, product := range dataFile.Products {
		dbStock := DBStock{
			ProductID: product.ID,
			Count:     *defaultStock,
		}

		av, err := attributevalue.MarshalMap(dbStock)
		if err != nil {
			log.Fatalf("Failed to marshal stock for product %s: %v", product.ID, err)
		}

		_, err = client.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(*stocksTable),
			Item:      av,
		})
		if err != nil {
			log.Fatalf("Failed to insert stock for product %s: %v", product.ID, err)
		}

		log.Printf("Inserted stock: product_id=%s, count=%d", product.ID, *defaultStock)
	}

	log.Printf("Successfully populated %s and %s with %d products", *productsTable, *stocksTable, len(dataFile.Products))
}

func clearTable(ctx context.Context, client *dynamodb.Client, tableName string) error {
	// Scan all items
	result, err := client.Scan(ctx, &dynamodb.ScanInput{
		TableName: aws.String(tableName),
	})
	if err != nil {
		return fmt.Errorf("scan table: %w", err)
	}

	// Delete each item
	for _, item := range result.Items {
		// Build key from the item
		key := make(map[string]types.AttributeValue)

		// Try to get 'id' key (for products table)
		if v, ok := item["id"]; ok {
			key["id"] = v
		}

		// Try to get 'product_id' key (for stocks table)
		if v, ok := item["product_id"]; ok {
			key["product_id"] = v
		}

		if len(key) == 0 {
			continue
		}

		_, err = client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
			TableName: aws.String(tableName),
			Key:       key,
		})
		if err != nil {
			return fmt.Errorf("delete item: %w", err)
		}
	}

	return nil
}
