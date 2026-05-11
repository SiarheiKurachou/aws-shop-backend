package importfileparser

import (
	"context"
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
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

	expected := []map[string]string{
		{"title": "Book", "price": "10", "count": "3"},
		{"title": "Pen", "price": "2", "count": ""},
		{"title": "Notebook", "price": "7", "count": "15", "extra_1": "unexpected"},
	}

	if !reflect.DeepEqual(expected, records) {
		t.Fatalf("expected records %+v, got %+v", expected, records)
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
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		newS3Client = originalNewS3Client
	})

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	mockClient := &mockS3Client{
		getObjectOutput: &s3.GetObjectOutput{
			Body: io.NopCloser(strings.NewReader("title,price\nBook,10\n")),
		},
	}
	newS3Client = func(aws.Config) s3ObjectAPI {
		return mockClient
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
}
