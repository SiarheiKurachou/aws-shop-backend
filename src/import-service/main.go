package importservice

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const signedURLExpiry = 5 * time.Minute

func HandleImportProductsFile(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	fileName := strings.TrimSpace(event.QueryStringParameters["name"])
	if fileName == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusBadRequest,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type":                "text/plain",
			},
			Body: "name query parameter is required",
		}, nil
	}

	bucketName := strings.TrimSpace(os.Getenv("IMPORT_BUCKET_NAME"))
	if bucketName == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type":                "text/plain",
			},
			Body: "IMPORT_BUCKET_NAME is not configured",
		}, nil
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type":                "text/plain",
			},
			Body: "failed to load AWS config",
		}, nil
	}

	presignClient := s3.NewPresignClient(s3.NewFromConfig(cfg))
	presignedReq, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: &bucketName,
		Key:    awsString(fmt.Sprintf("uploaded/%s", fileName)),
	}, func(options *s3.PresignOptions) {
		options.Expires = signedURLExpiry
	})
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: http.StatusInternalServerError,
			Headers: map[string]string{
				"Access-Control-Allow-Origin": "*",
				"Content-Type":                "text/plain",
			},
			Body: "failed to generate signed URL",
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Origin": "*",
			"Content-Type":                "text/plain",
		},
		Body: presignedReq.URL,
	}, nil
}

func awsString(v string) *string {
	return &v
}
