package importproductsfile

import (
	"context"
	"net/http"
	"os"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

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
