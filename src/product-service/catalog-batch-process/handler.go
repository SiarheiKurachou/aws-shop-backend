package catalogbatchprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"aws-shop-backend/src/product-service/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

var createProduct = core.CreateProduct
var loadAWSConfig = config.LoadDefaultConfig

type snsPublishAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

var newSNSClient = func(cfg aws.Config) snsPublishAPI {
	return sns.NewFromConfig(cfg)
}

// HandleCatalogBatchProcess processes SQS records and creates products for each message.
func HandleCatalogBatchProcess(ctx context.Context, event events.SQSEvent) error {
	createdProducts := make([]core.Product, 0, len(event.Records))

	for _, message := range event.Records {
		var req core.CreateProductRequest
		if err := json.Unmarshal([]byte(message.Body), &req); err != nil {
			return fmt.Errorf("unmarshal SQS message %s: %w", message.MessageId, err)
		}

		product, err := createProduct(req)
		if err != nil {
			return fmt.Errorf("create product from message %s: %w", message.MessageId, err)
		}

		createdProducts = append(createdProducts, product)

		log.Printf("created product from SQS message: %s", message.MessageId)
	}

	if len(createdProducts) == 0 {
		return nil
	}

	topicARN := strings.TrimSpace(os.Getenv("CREATE_PRODUCT_TOPIC_ARN"))
	if topicARN == "" {
		return fmt.Errorf("CREATE_PRODUCT_TOPIC_ARN is not configured")
	}

	cfg, err := loadAWSConfig(ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	snsClient := newSNSClient(cfg)

	eventPayload, err := json.Marshal(map[string]any{
		"count":    len(createdProducts),
		"products": createdProducts,
	})
	if err != nil {
		return fmt.Errorf("marshal create product event: %w", err)
	}

	_, err = snsClient.Publish(ctx, &sns.PublishInput{
		TopicArn: aws.String(topicARN),
		Subject:  aws.String("Products created"),
		Message:  aws.String(string(eventPayload)),
		MessageAttributes: map[string]snstypes.MessageAttributeValue{
			"priceCategory": {
				DataType:    aws.String("String"),
				StringValue: aws.String(priceCategory(createdProducts)),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("publish create product event: %w", err)
	}

	return nil
}

func priceCategory(products []core.Product) string {
	for _, product := range products {
		if product.Price >= 100 {
			return "premium"
		}
	}

	return "budget"
}
