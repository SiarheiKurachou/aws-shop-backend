package catalogbatchprocess

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"aws-shop-backend/src/product-service/core"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

type mockSNSClient struct {
	publishInput *sns.PublishInput
	publishErr   error
}

func (m *mockSNSClient) Publish(_ context.Context, params *sns.PublishInput, _ ...func(*sns.Options)) (*sns.PublishOutput, error) {
	m.publishInput = params
	if m.publishErr != nil {
		return nil, m.publishErr
	}

	return &sns.PublishOutput{}, nil
}

func TestHandleCatalogBatchProcessSuccess(t *testing.T) {
	originalCreateProduct := createProduct
	originalLoadAWSConfig := loadAWSConfig
	originalNewSNSClient := newSNSClient
	originalTopicARN := os.Getenv("CREATE_PRODUCT_TOPIC_ARN")
	t.Cleanup(func() {
		createProduct = originalCreateProduct
		loadAWSConfig = originalLoadAWSConfig
		newSNSClient = originalNewSNSClient
		_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", originalTopicARN)
	})

	requests := make([]core.CreateProductRequest, 0)
	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		requests = append(requests, req)
		return core.Product{ID: "p-id", Title: req.Title}, nil
	}

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	mockSNS := &mockSNSClient{}
	newSNSClient = func(aws.Config) snsPublishAPI {
		return mockSNS
	}
	_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", "arn:aws:sns:us-east-1:123:createProductTopic")

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "m1", Body: `{"title":"A","description":"d1","price":10,"count":1}`},
			{MessageId: "m2", Body: `{"title":"B","description":"d2","price":20,"count":2}`},
		},
	}

	err := HandleCatalogBatchProcess(context.Background(), event)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(requests) != 2 {
		t.Fatalf("expected 2 createProduct calls, got: %d", len(requests))
	}

	if requests[0].Title != "A" || requests[1].Title != "B" {
		t.Fatalf("unexpected request sequence: %#v", requests)
	}

	if mockSNS.publishInput == nil {
		t.Fatal("expected SNS publish call")
	}

	if got := aws.ToString(mockSNS.publishInput.TopicArn); got != "arn:aws:sns:us-east-1:123:createProductTopic" {
		t.Fatalf("unexpected topic arn: %s", got)
	}

	message := aws.ToString(mockSNS.publishInput.Message)
	if !strings.Contains(message, `"count":2`) {
		t.Fatalf("expected count in SNS message, got: %s", message)
	}

	if mockSNS.publishInput.MessageAttributes == nil {
		t.Fatal("expected message attributes")
	}

	priceCategoryAttr, exists := mockSNS.publishInput.MessageAttributes["priceCategory"]
	if !exists {
		t.Fatal("expected priceCategory message attribute")
	}

	if got := aws.ToString(priceCategoryAttr.StringValue); got != "budget" {
		t.Fatalf("expected budget price category, got: %s", got)
	}
}

func TestHandleCatalogBatchProcessPremiumCategory(t *testing.T) {
	originalCreateProduct := createProduct
	originalLoadAWSConfig := loadAWSConfig
	originalNewSNSClient := newSNSClient
	originalTopicARN := os.Getenv("CREATE_PRODUCT_TOPIC_ARN")
	t.Cleanup(func() {
		createProduct = originalCreateProduct
		loadAWSConfig = originalLoadAWSConfig
		newSNSClient = originalNewSNSClient
		_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", originalTopicARN)
	})

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{ID: "p-id", Price: 250}, nil
	}

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	mockSNS := &mockSNSClient{}
	newSNSClient = func(aws.Config) snsPublishAPI {
		return mockSNS
	}
	_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", "arn:aws:sns:us-east-1:123:createProductTopic")

	err := HandleCatalogBatchProcess(context.Background(), events.SQSEvent{
		Records: []events.SQSMessage{{MessageId: "m1", Body: `{"title":"A","description":"d1","price":250,"count":1}`}},
	})
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	priceCategoryAttr, exists := mockSNS.publishInput.MessageAttributes["priceCategory"]
	if !exists {
		t.Fatal("expected priceCategory message attribute")
	}

	if got := aws.ToString(priceCategoryAttr.StringValue); got != "premium" {
		t.Fatalf("expected premium price category, got: %s", got)
	}
}

func TestHandleCatalogBatchProcessInvalidJSON(t *testing.T) {
	originalCreateProduct := createProduct
	originalTopicARN := os.Getenv("CREATE_PRODUCT_TOPIC_ARN")
	t.Cleanup(func() {
		createProduct = originalCreateProduct
		_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", originalTopicARN)
	})

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{}, nil
	}

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "bad-message", Body: `{"title":"A"`},
		},
	}

	err := HandleCatalogBatchProcess(context.Background(), event)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unmarshal SQS message bad-message") {
		t.Fatalf("expected unmarshal error with message id, got: %v", err)
	}
}

func TestHandleCatalogBatchProcessCreateProductError(t *testing.T) {
	originalCreateProduct := createProduct
	originalTopicARN := os.Getenv("CREATE_PRODUCT_TOPIC_ARN")
	t.Cleanup(func() {
		createProduct = originalCreateProduct
		_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", originalTopicARN)
	})

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{}, errors.New("boom")
	}

	event := events.SQSEvent{
		Records: []events.SQSMessage{
			{MessageId: "m-fail", Body: `{"title":"A","description":"d1","price":10,"count":1}`},
		},
	}

	err := HandleCatalogBatchProcess(context.Background(), event)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "create product from message m-fail") {
		t.Fatalf("expected create product error with message id, got: %v", err)
	}
}

func TestHandleCatalogBatchProcessMissingTopicARN(t *testing.T) {
	originalCreateProduct := createProduct
	originalTopicARN := os.Getenv("CREATE_PRODUCT_TOPIC_ARN")
	t.Cleanup(func() {
		createProduct = originalCreateProduct
		_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", originalTopicARN)
	})

	_ = os.Unsetenv("CREATE_PRODUCT_TOPIC_ARN")

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{ID: "p-id"}, nil
	}

	err := HandleCatalogBatchProcess(context.Background(), events.SQSEvent{
		Records: []events.SQSMessage{{MessageId: "m1", Body: `{"title":"A","description":"d1","price":10,"count":1}`}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "CREATE_PRODUCT_TOPIC_ARN is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleCatalogBatchProcessPublishError(t *testing.T) {
	originalCreateProduct := createProduct
	originalLoadAWSConfig := loadAWSConfig
	originalNewSNSClient := newSNSClient
	originalTopicARN := os.Getenv("CREATE_PRODUCT_TOPIC_ARN")
	t.Cleanup(func() {
		createProduct = originalCreateProduct
		loadAWSConfig = originalLoadAWSConfig
		newSNSClient = originalNewSNSClient
		_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", originalTopicARN)
	})

	_ = os.Setenv("CREATE_PRODUCT_TOPIC_ARN", "arn:aws:sns:us-east-1:123:createProductTopic")

	createProduct = func(req core.CreateProductRequest) (core.Product, error) {
		return core.Product{ID: "p-id"}, nil
	}

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	newSNSClient = func(aws.Config) snsPublishAPI {
		return &mockSNSClient{publishErr: errors.New("publish failed")}
	}

	err := HandleCatalogBatchProcess(context.Background(), events.SQSEvent{
		Records: []events.SQSMessage{{MessageId: "m1", Body: `{"title":"A","description":"d1","price":10,"count":1}`}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "publish create product event") {
		t.Fatalf("unexpected error: %v", err)
	}
}
