package importservice

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/url"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func HandleImportFileParser(ctx context.Context, event events.S3Event) error {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	for _, record := range event.Records {
		bucketName := record.S3.Bucket.Name
		if bucketName == "" {
			log.Printf("skip record without bucket: %+v", record)
			continue
		}

		decodedKey, err := url.QueryUnescape(record.S3.Object.Key)
		if err != nil {
			return fmt.Errorf("decode object key %q: %w", record.S3.Object.Key, err)
		}

		if decodedKey == "" {
			log.Printf("skip record without key: %+v", record)
			continue
		}

		obj, err := s3Client.GetObject(ctx, &s3.GetObjectInput{
			Bucket: awsString(bucketName),
			Key:    awsString(decodedKey),
		})
		if err != nil {
			return fmt.Errorf("get s3 object %s/%s: %w", bucketName, decodedKey, err)
		}

		csvReader := csv.NewReader(obj.Body)
		headers, readErr := csvReader.Read()
		if readErr == io.EOF {
			if closeErr := obj.Body.Close(); closeErr != nil {
				return fmt.Errorf("close body %s/%s: %w", bucketName, decodedKey, closeErr)
			}
			continue
		}
		if readErr != nil {
			_ = obj.Body.Close()
			return fmt.Errorf("read headers %s/%s: %w", bucketName, decodedKey, readErr)
		}

		for {
			recordValues, readErr := csvReader.Read()
			if readErr == io.EOF {
				break
			}
			if readErr != nil {
				_ = obj.Body.Close()
				return fmt.Errorf("parse csv %s/%s: %w", bucketName, decodedKey, readErr)
			}

			recordMap := map[string]string{}
			for i, header := range headers {
				if i >= len(recordValues) {
					recordMap[header] = ""
					continue
				}
				recordMap[header] = recordValues[i]
			}

			if len(recordValues) > len(headers) {
				for i := len(headers); i < len(recordValues); i++ {
					recordMap[fmt.Sprintf("extra_%d", i-len(headers)+1)] = recordValues[i]
				}
			}

			log.Printf("CSV record: %v", recordMap)
		}

		if closeErr := obj.Body.Close(); closeErr != nil {
			return fmt.Errorf("close body %s/%s: %w", bucketName, decodedKey, closeErr)
		}
	}

	return nil
}
