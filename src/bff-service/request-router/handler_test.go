package requestrouter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleProxyRequestReturns502WhenRecipientIsUnknown(t *testing.T) {
	t.Setenv("CART_URL", "")
	defer func(oldClient *http.Client) {
		proxyHTTPClient = oldClient
	}(proxyHTTPClient)
	proxyHTTPClient = http.DefaultClient

	req := httptest.NewRequest(http.MethodGet, "/unknown?var1=someValue", nil)
	rr := httptest.NewRecorder()

	HandleProxyRequest(rr, req)

	if rr.Code != http.StatusBadGateway {
		t.Fatalf("expected status %d, got %d", http.StatusBadGateway, rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "Cannot process request") {
		t.Fatalf("expected Cannot process request message, got %s", rr.Body.String())
	}
}

func TestHandleProxyRequestForwardsMethodQueryHeadersAndBody(t *testing.T) {
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected method POST, got %s", r.Method)
		}

		if r.URL.Path != "/orders" {
			t.Fatalf("expected path /orders, got %s", r.URL.Path)
		}

		if r.URL.Query().Get("var1") != "someValue" {
			t.Fatalf("expected query var1=someValue, got %s", r.URL.RawQuery)
		}

		if r.Header.Get("X-Test") != "header-value" {
			t.Fatalf("expected X-Test header to be forwarded")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("failed to read request body: %v", err)
		}

		if string(body) != "sample-body" {
			t.Fatalf("expected body sample-body, got %s", string(body))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	defer upstreamServer.Close()

	t.Setenv("CART_URL", upstreamServer.URL)
	defer func(oldClient *http.Client) {
		proxyHTTPClient = oldClient
	}(proxyHTTPClient)
	proxyHTTPClient = upstreamServer.Client()

	req := httptest.NewRequest(http.MethodPost, "/cart/orders?var1=someValue", strings.NewReader("sample-body"))
	req.Header.Set("X-Test", "header-value")
	rr := httptest.NewRecorder()

	HandleProxyRequest(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rr.Code)
	}

	if rr.Body.String() != `{"ok":true}` {
		t.Fatalf("expected response body from upstream, got %s", rr.Body.String())
	}
}

func TestHandleProxyRequestPassesThroughUpstreamError(t *testing.T) {
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("upstream not found"))
	}))
	defer upstreamServer.Close()

	t.Setenv("PRODUCT_URL", upstreamServer.URL)
	defer func(oldClient *http.Client) {
		proxyHTTPClient = oldClient
	}(proxyHTTPClient)
	proxyHTTPClient = upstreamServer.Client()

	req := httptest.NewRequest(http.MethodGet, "/product?var1=someValue", nil)
	rr := httptest.NewRecorder()

	HandleProxyRequest(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}

	if rr.Body.String() != "upstream not found" {
		t.Fatalf("expected upstream error body to pass through, got %s", rr.Body.String())
	}
}

func TestResolveRecipientURLUsesSupportedKeys(t *testing.T) {
	t.Setenv("CART_URL", "http://example.local")

	value := resolveRecipientURL("cart")
	if value != "http://example.local" {
		t.Fatalf("expected CART_URL to be used, got %s", value)
	}
}

func TestResolveRecipientURLRejectsUnsupportedService(t *testing.T) {
	t.Setenv("PAYMENT_URL", "http://example.local")

	value := resolveRecipientURL("payment")
	if value != "" {
		t.Fatalf("expected unsupported service to return empty URL, got %s", value)
	}
}
