package importfileparser

import (
	"context"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type mockS3Client struct {
	getObjectOutput *s3.GetObjectOutput
	getObjectErr    error
	copyObjectErr   error
	deleteObjectErr error

	getObjectInput    *s3.GetObjectInput
	copyObjectInput   *s3.CopyObjectInput
	deleteObjectInput *s3.DeleteObjectInput
}

type mockSQSClient struct {
	sendMessageInputs []*sqs.SendMessageInput
	sendMessageErr    error
}

func (m *mockSQSClient) SendMessage(_ context.Context, params *sqs.SendMessageInput, _ ...func(*sqs.Options)) (*sqs.SendMessageOutput, error) {
	m.sendMessageInputs = append(m.sendMessageInputs, params)
	if m.sendMessageErr != nil {
		return nil, m.sendMessageErr
	}
	return &sqs.SendMessageOutput{}, nil
}

func (m *mockS3Client) GetObject(_ context.Context, params *s3.GetObjectInput, _ ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	m.getObjectInput = params
	if m.getObjectErr != nil {
		return nil, m.getObjectErr
	}
	return m.getObjectOutput, nil
}

func (m *mockS3Client) CopyObject(_ context.Context, params *s3.CopyObjectInput, _ ...func(*s3.Options)) (*s3.CopyObjectOutput, error) {
	m.copyObjectInput = params
	if m.copyObjectErr != nil {
		return nil, m.copyObjectErr
	}
	return &s3.CopyObjectOutput{}, nil
}

func (m *mockS3Client) DeleteObject(_ context.Context, params *s3.DeleteObjectInput, _ ...func(*s3.Options)) (*s3.DeleteObjectOutput, error) {
	m.deleteObjectInput = params
	if m.deleteObjectErr != nil {
		return nil, m.deleteObjectErr
	}
	return &s3.DeleteObjectOutput{}, nil
}

func TestParseCSVRecordsWithHeaders(t *testing.T) {
	input := strings.NewReader("title,price,count\nBook,10,3\nPen,2\nNotebook,7,15,unexpected\n")

	records, err := parseCSVRecordsWithHeaders(input)
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	expected := []map[string]any{
		{"title": "Book", "price": 10, "count": 3},
		{"title": "Pen", "price": 2},
		{"title": "Notebook", "price": 7, "count": 15, "extra_1": "unexpected"},
	}

	if !reflect.DeepEqual(expected, records) {
		t.Fatalf("expected records %+v, got %+v", expected, records)
	}
}

func TestParseCSVRecordsWithHeaders_InvalidNumericField(t *testing.T) {
	input := strings.NewReader("title,price,count\nBook,abc,3\n")

	_, err := parseCSVRecordsWithHeaders(input)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "parse price in row 1") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleImportFileParser_EmptyEvent(t *testing.T) {
	err := HandleImportFileParser(context.Background(), events.S3Event{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}

func TestHandleImportFileParser_MovesFileToParsedPrefix(t *testing.T) {
	originalLoadAWSConfig := loadAWSConfig
	originalNewS3Client := newS3Client
	originalNewSQSClient := newSQSClient
	originalQueueURL := os.Getenv("CATALOG_ITEMS_QUEUE_URL")
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		newS3Client = originalNewS3Client
		newSQSClient = originalNewSQSClient
		_ = os.Setenv("CATALOG_ITEMS_QUEUE_URL", originalQueueURL)
	})

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	mockClient := &mockS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader("title,price\nBook,10\n")),
		},
	}
	mockSQS := &mockSQSClient{}
	newS3Client = func(aws.Config) s3ObjectAPI {
		return mockClient
	}
	newSQSClient = func(aws.Config) sqsSendMessageAPI {
		return mockSQS
	}
	_ = os.Setenv("CATALOG_ITEMS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/123/catalogItemsQueue")

	err := HandleImportFileParser(context.Background(), events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "import-bucket"},
					Object: events.S3Object{Key: "uploaded/products.csv"},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got := aws.ToString(mockClient.getObjectInput.Key); got != "uploaded/products.csv" {
		t.Fatalf("expected get key uploaded/products.csv, got %q", got)
	}

	if got := aws.ToString(mockClient.copyObjectInput.Key); got != "parsed/products.csv" {
		t.Fatalf("expected copy destination parsed/products.csv, got %q", got)
	}

	if got := aws.ToString(mockClient.copyObjectInput.CopySource); got != "import-bucket/uploaded/products.csv" {
		t.Fatalf("expected copy source import-bucket/uploaded/products.csv, got %q", got)
	}

	if got := aws.ToString(mockClient.deleteObjectInput.Key); got != "uploaded/products.csv" {
		t.Fatalf("expected delete key uploaded/products.csv, got %q", got)
	}

	if len(mockSQS.sendMessageInputs) != 1 {
		t.Fatalf("expected 1 SQS message, got %d", len(mockSQS.sendMessageInputs))
	}

	if got := aws.ToString(mockSQS.sendMessageInputs[0].QueueUrl); got != "https://sqs.us-east-1.amazonaws.com/123/catalogItemsQueue" {
		t.Fatalf("expected queue URL to match env, got %q", got)
	}

	if got := aws.ToString(mockSQS.sendMessageInputs[0].MessageBody); got != `{"price":10,"title":"Book"}` && got != `{"title":"Book","price":10}` {
		t.Fatalf("unexpected message body: %s", got)
	}
}

func TestHandleImportFileParser_MissingQueueURL(t *testing.T) {
	originalLoadAWSConfig := loadAWSConfig
	originalQueueURL := os.Getenv("CATALOG_ITEMS_QUEUE_URL")
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		_ = os.Setenv("CATALOG_ITEMS_QUEUE_URL", originalQueueURL)
	})

	_ = os.Unsetenv("CATALOG_ITEMS_QUEUE_URL")

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	err := HandleImportFileParser(context.Background(), events.S3Event{
		Records: []events.S3EventRecord{{}},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "CATALOG_ITEMS_QUEUE_URL is not configured") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHandleImportFileParser_SQSFailure(t *testing.T) {
	originalLoadAWSConfig := loadAWSConfig
	originalNewS3Client := newS3Client
	originalNewSQSClient := newSQSClient
	originalQueueURL := os.Getenv("CATALOG_ITEMS_QUEUE_URL")
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		newS3Client = originalNewS3Client
		newSQSClient = originalNewSQSClient
		_ = os.Setenv("CATALOG_ITEMS_QUEUE_URL", originalQueueURL)
	})

	_ = os.Setenv("CATALOG_ITEMS_QUEUE_URL", "https://sqs.us-east-1.amazonaws.com/123/catalogItemsQueue")
	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	newS3Client = func(aws.Config) s3ObjectAPI {
		return &mockS3Client{
			getObjectOutput: &s3.GetObjectOutput{
				Body: io.NopCloser(strings.NewReader("title,price\nBook,10\n")),
			},
		}
	}

	newSQSClient = func(aws.Config) sqsSendMessageAPI {
		return &mockSQSClient{sendMessageErr: errors.New("send failed")}
	}

	err := HandleImportFileParser(context.Background(), events.S3Event{
		Records: []events.S3EventRecord{
			{
				S3: events.S3Entity{
					Bucket: events.S3Bucket{Name: "import-bucket"},
					Object: events.S3Object{Key: "uploaded/products.csv"},
				},
			},
		},
	})
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "send csv record to SQS") {
		t.Fatalf("unexpected error: %v", err)
	}
}
