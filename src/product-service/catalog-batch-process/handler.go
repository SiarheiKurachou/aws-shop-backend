package catalogbatchprocess

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"aws-shop-backend/src/product-service/common"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	snstypes "github.com/aws/aws-sdk-go-v2/service/sns/types"
)

var createProduct = common.CreateProduct
var loadAWSConfig = config.LoadDefaultConfig

type snsPublishAPI interface {
	Publish(ctx context.Context, params *sns.PublishInput, optFns ...func(*sns.Options)) (*sns.PublishOutput, error)
}

var newSNSClient = func(cfg aws.Config) snsPublishAPI {
	return sns.NewFromConfig(cfg)
}

// HandleCatalogBatchProcess processes SQS records and creates products for each message.
func HandleCatalogBatchProcess(ctx context.Context, event events.SQSEvent) error {
	createdProducts := make([]common.Product, 0, len(event.Records))

	for _, message := range event.Records {
		req, err := decodeCreateProductRequest(message.Body)
		if err != nil {
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

type intStringNumber int

func (v *intStringNumber) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var asInt int
	if err := json.Unmarshal(data, &asInt); err == nil {
		*v = intStringNumber(asInt)
		return nil
	}

	var asString string
	if err := json.Unmarshal(data, &asString); err != nil {
		return fmt.Errorf("expected integer or numeric string")
	}

	asString = strings.TrimSpace(asString)
	if asString == "" {
		return fmt.Errorf("empty numeric string")
	}

	parsedInt, err := strconv.Atoi(asString)
	if err != nil {
		return fmt.Errorf("parse int %q: %w", asString, err)
	}

	*v = intStringNumber(parsedInt)
	return nil
}

type createProductRequestMessage struct {
	ID          string          `json:"id,omitempty"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	Price       intStringNumber `json:"price"`
	Count       intStringNumber `json:"count,omitempty"`
}

func decodeCreateProductRequest(body string) (common.CreateProductRequest, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal([]byte(body), &raw); err != nil {
		return common.CreateProductRequest{}, err
	}

	normalized := make(map[string]json.RawMessage, len(raw))
	for key, value := range raw {
		normalized[normalizeJSONKey(key)] = value
	}

	var message createProductRequestMessage
	if err := unmarshalField(normalized, "id", &message.ID); err != nil {
		return common.CreateProductRequest{}, fmt.Errorf("id: %w", err)
	}
	if err := unmarshalField(normalized, "title", &message.Title); err != nil {
		return common.CreateProductRequest{}, fmt.Errorf("title: %w", err)
	}
	if err := unmarshalField(normalized, "description", &message.Description); err != nil {
		return common.CreateProductRequest{}, fmt.Errorf("description: %w", err)
	}
	if err := unmarshalField(normalized, "price", &message.Price); err != nil {
		return common.CreateProductRequest{}, fmt.Errorf("price: %w", err)
	}
	if err := unmarshalField(normalized, "count", &message.Count); err != nil {
		return common.CreateProductRequest{}, fmt.Errorf("count: %w", err)
	}

	return common.CreateProductRequest{
		ID:          message.ID,
		Title:       strings.TrimSpace(message.Title),
		Description: strings.TrimSpace(message.Description),
		Price:       int(message.Price),
		Count:       int(message.Count),
	}, nil
}

func unmarshalField[T any](source map[string]json.RawMessage, key string, out *T) error {
	rawValue, exists := source[key]
	if !exists {
		return nil
	}

	if err := json.Unmarshal(rawValue, out); err != nil {
		return err
	}

	return nil
}

func normalizeJSONKey(key string) string {
	cleaned := strings.TrimSpace(key)
	cleaned = strings.TrimPrefix(cleaned, "\ufeff")
	return strings.ToLower(cleaned)
}

func priceCategory(products []common.Product) string {
	for _, product := range products {
		if product.Price >= 100 {
			return "premium"
		}
	}

	return "budget"
}
