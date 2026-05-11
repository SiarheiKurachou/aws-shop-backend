package importproductsfile

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type mockPresignClient struct {
	request *v4.PresignedHTTPRequest
	err     error
	input   *s3.PutObjectInput
}

func (m *mockPresignClient) PresignPutObject(_ context.Context, input *s3.PutObjectInput, _ ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error) {
	m.input = input
	if m.err != nil {
		return nil, m.err
	}
	return m.request, nil
}

func TestHandleImportProductsFile_MissingName(t *testing.T) {
	t.Setenv("IMPORT_BUCKET_NAME", "test-bucket")

	response, err := HandleImportProductsFile(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, response.StatusCode)
	}
}

func TestHandleImportProductsFile_MissingBucketEnv(t *testing.T) {
	if err := os.Unsetenv("IMPORT_BUCKET_NAME"); err != nil {
		t.Fatalf("failed to unset env: %v", err)
	}

	response, err := HandleImportProductsFile(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"name": "products.csv"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, response.StatusCode)
	}
}

func TestHandleImportProductsFile_ReturnsSignedURL(t *testing.T) {
	t.Setenv("IMPORT_BUCKET_NAME", "import-bucket")

	originalLoadAWSConfig := loadAWSConfig
	originalNewPresignClient := newPresignClient
	t.Cleanup(func() {
		loadAWSConfig = originalLoadAWSConfig
		newPresignClient = originalNewPresignClient
	})

	loadAWSConfig = func(context.Context, ...func(*config.LoadOptions) error) (aws.Config, error) {
		return aws.Config{}, nil
	}

	mockClient := &mockPresignClient{
		request: &v4.PresignedHTTPRequest{URL: "https://signed.example.com/upload"},
	}
	newPresignClient = func(aws.Config) presignPutObjectAPI {
		return mockClient
	}

	response, err := HandleImportProductsFile(context.Background(), events.APIGatewayProxyRequest{
		QueryStringParameters: map[string]string{"name": "products.csv"},
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.StatusCode)
	}

	if response.Body != "https://signed.example.com/upload" {
		t.Fatalf("expected signed URL in body, got %q", response.Body)
	}

	if mockClient.input == nil {
		t.Fatalf("expected PresignPutObject to receive input")
	}

	if got, want := aws.ToString(mockClient.input.Bucket), "import-bucket"; got != want {
		t.Fatalf("expected bucket %q, got %q", want, got)
	}

	if got, want := aws.ToString(mockClient.input.Key), "uploaded/products.csv"; got != want {
		t.Fatalf("expected key %q, got %q", want, got)
	}
}
