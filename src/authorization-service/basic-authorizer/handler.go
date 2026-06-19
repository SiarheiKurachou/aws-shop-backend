package basicauthorizer

import (
	"context"
	"encoding/base64"
	"errors"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

const unauthorizedError = "Unauthorized"

// HandleBasicAuthorizer validates basic auth credentials against Lambda env vars.
// @Summary      Validate basic authorization token
// @Description  Decodes Basic token, validates credentials from Lambda environment variables, and returns an IAM policy.
// @Tags         authorization
// @Accept       json
// @Produce      json
// @Param        Authorization  header  string  true  "Basic token"
// @Success      200  {object}  map[string]interface{}  "Allow policy"
// @Failure      401  {string}  string  "Authorization header is missing"
// @Failure      403  {string}  string  "Invalid token or credentials"
// @Router       /auth [get]
func HandleBasicAuthorizer(_ context.Context, event events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token := strings.TrimSpace(event.AuthorizationToken)
	if token == "" {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New(unauthorizedError)
	}

	username, password, ok := decodeBasicCredentials(token)
	if !ok {
		return buildPolicy("anonymous", "Deny", event.MethodArn), nil
	}

	if isAuthorized(username, password) {
		return buildPolicy(username, "Allow", event.MethodArn), nil
	}

	return buildPolicy(username, "Deny", event.MethodArn), nil
}

func decodeBasicCredentials(token string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(token), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Basic") {
		return "", "", false
	}

	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(parts[1]))
	if err != nil {
		return "", "", false
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return "", "", false
	}

	username := strings.TrimSpace(credentials[0])
	if username == "" {
		return "", "", false
	}

	return username, credentials[1], true
}

func isAuthorized(username string, password string) bool {
	expectedPassword := os.Getenv(username)
	return expectedPassword != "" && expectedPassword == password
}

func buildPolicy(principalID string, effect string, methodArn string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{
				{
					Action:   []string{"execute-api:Invoke"},
					Effect:   effect,
					Resource: []string{methodArn},
				},
			},
		},
	}
}
