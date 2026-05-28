package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	basicauthorizer "aws-shop-backend/src/authorization-service/basic-authorizer"

	"github.com/aws/aws-lambda-go/events"
)

// @title AWS Shop Backend Authorization API
// @version 1.0
// @description Authorization service API
// @BasePath /

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		authorizerResponse, err := basicauthorizer.HandleBasicAuthorizer(context.Background(), events.APIGatewayCustomAuthorizerRequest{
			AuthorizationToken: r.Header.Get("Authorization"),
			MethodArn:          "arn:aws:execute-api:local:000000000000:api/prod/GET/import",
		})
		if err != nil {
			if err.Error() == "Unauthorized" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(err.Error()))
			return
		}

		effect := ""
		if len(authorizerResponse.PolicyDocument.Statement) > 0 {
			effect = authorizerResponse.PolicyDocument.Statement[0].Effect
		}

		if effect == "Deny" {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		responseBody, marshalErr := json.Marshal(authorizerResponse)
		if marshalErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(marshalErr.Error()))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(responseBody)
	})

	log.Println("Authorization service server starting on :8082")
	log.Fatal(http.ListenAndServe(":8082", mux))
}
