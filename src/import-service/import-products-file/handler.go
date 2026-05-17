package importproductsfile

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const signedURLExpiry = 5 * time.Minute

type presignPutObjectAPI interface {
	PresignPutObject(ctx context.Context, params *s3.PutObjectInput, optFns ...func(*s3.PresignOptions)) (*v4.PresignedHTTPRequest, error)
}

var loadAWSConfig = config.LoadDefaultConfig
var newPresignClient = func(cfg aws.Config) presignPutObjectAPI {
	return s3.NewPresignClient(s3.NewFromConfig(cfg))
}

// HandleImportProductsFile generates a presigned S3 URL for file upload.
//
// @Summary      Get presigned S3 upload URL
// @Description  Returns a presigned URL to upload a file to S3.
// @Tags         import
// @Accept       json
// @Produce      json
// @Param        name  query  string  true  "File name"
// @Success      200   {string}  string  "Presigned URL"
// @Failure      400   {string}  string  "Missing file name"
// @Failure      500   {string}  string  "Server error"
// @Router       /import [get]
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

	cfg, err := loadAWSConfig(ctx)
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

	presignClient := newPresignClient(cfg)
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
