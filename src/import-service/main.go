package main

import (
	"context"
	"log"
	"net/http"

	importproductsfile "aws-shop-backend/src/import-service/import-products-file"

	"github.com/aws/aws-lambda-go/events"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/import", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		event := events.APIGatewayProxyRequest{
			QueryStringParameters: map[string]string{
				"name": r.URL.Query().Get("name"),
			},
		}

		response, err := importproductsfile.HandleImportProductsFile(context.Background(), event)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}
		w.WriteHeader(response.StatusCode)
		_, _ = w.Write([]byte(response.Body))
	})

	log.Println("Import service server starting on :8081")
	log.Fatal(http.ListenAndServe(":8081", mux))
}
