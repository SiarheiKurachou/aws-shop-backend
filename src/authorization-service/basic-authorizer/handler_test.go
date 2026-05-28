package basicauthorizer

import (
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

func TestHandleBasicAuthorizer_MissingAuthorizationHeader(t *testing.T) {
	response, err := HandleBasicAuthorizer(context.Background(), events.APIGatewayCustomAuthorizerRequest{
		MethodArn: "arn:aws:execute-api:eu-west-1:123456789012:api-id/prod/GET/import",
	})

	if err == nil {
		t.Fatal("expected unauthorized error, got nil")
	}

	if err.Error() != unauthorizedError {
		t.Fatalf("expected %q error, got %q", unauthorizedError, err.Error())
	}

	if response.PolicyDocument.Version != "" {
		t.Fatalf("expected empty response on unauthorized error, got %#v", response)
	}
}

func TestHandleBasicAuthorizer_InvalidToken_ReturnsDenyPolicy(t *testing.T) {
	t.Setenv("siarhei_kurachou", "TEST_PASSWORD")

	response, err := HandleBasicAuthorizer(context.Background(), events.APIGatewayCustomAuthorizerRequest{
		AuthorizationToken: "Basic invalid-base64",
		MethodArn:          "arn:aws:execute-api:eu-west-1:123456789012:api-id/prod/GET/import",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got, want := response.PolicyDocument.Statement[0].Effect, "Deny"; got != want {
		t.Fatalf("expected effect %q, got %q", want, got)
	}
}

func TestHandleBasicAuthorizer_ValidToken_ReturnsAllowPolicy(t *testing.T) {
	t.Setenv("siarhei_kurachou", "TEST_PASSWORD")

	authorizationToken := "Basic " + base64.StdEncoding.EncodeToString([]byte("siarhei_kurachou:TEST_PASSWORD"))
	response, err := HandleBasicAuthorizer(context.Background(), events.APIGatewayCustomAuthorizerRequest{
		AuthorizationToken: authorizationToken,
		MethodArn:          "arn:aws:execute-api:eu-west-1:123456789012:api-id/prod/GET/import",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got, want := response.PolicyDocument.Statement[0].Effect, "Allow"; got != want {
		t.Fatalf("expected effect %q, got %q", want, got)
	}

	if got, want := response.PrincipalID, "siarhei_kurachou"; got != want {
		t.Fatalf("expected principal %q, got %q", want, got)
	}

	if got, want := response.PolicyDocument.Statement[0].Action[0], "execute-api:Invoke"; got != want {
		t.Fatalf("expected action %q, got %q", want, got)
	}
}

func TestHandleBasicAuthorizer_WrongPassword_ReturnsDenyPolicy(t *testing.T) {
	t.Setenv("siarhei_kurachou", "TEST_PASSWORD")

	authorizationToken := "Basic " + base64.StdEncoding.EncodeToString([]byte("siarhei_kurachou:WRONG_PASSWORD"))
	response, err := HandleBasicAuthorizer(context.Background(), events.APIGatewayCustomAuthorizerRequest{
		AuthorizationToken: authorizationToken,
		MethodArn:          "arn:aws:execute-api:eu-west-1:123456789012:api-id/prod/GET/import",
	})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}

	if got, want := response.PolicyDocument.Statement[0].Effect, "Deny"; got != want {
		t.Fatalf("expected effect %q, got %q", want, got)
	}

	if response.PrincipalID == "" {
		t.Fatal("expected principal id to be set")
	}
}

func TestStatusCodesMatchRequirements(t *testing.T) {
	if got, want := http.StatusUnauthorized, 401; got != want {
		t.Fatalf("expected unauthorized status %d, got %d", want, got)
	}

	if got, want := http.StatusForbidden, 403; got != want {
		t.Fatalf("expected forbidden status %d, got %d", want, got)
	}
}
