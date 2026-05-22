package catalogbatchprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"aws-shop-backend/src/product-service/core"

	"github.com/aws/aws-lambda-go/events"
)

var createProduct = core.CreateProduct

// HandleCatalogBatchProcess processes SQS records and creates products for each message.
func HandleCatalogBatchProcess(_ context.Context, event events.SQSEvent) error {
	for _, message := range event.Records {
		var req core.CreateProductRequest
		if err := json.Unmarshal([]byte(message.Body), &req); err != nil {
			return fmt.Errorf("unmarshal SQS message %s: %w", message.MessageId, err)
		}

		if _, err := createProduct(req); err != nil {
			return fmt.Errorf("create product from message %s: %w", message.MessageId, err)
		}

		log.Printf("created product from SQS message: %s", message.MessageId)
	}

	return nil
}
