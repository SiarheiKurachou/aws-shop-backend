package requestrouter

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"aws-shop-backend/src/httphelpers"
)

var proxyHTTPClient = &http.Client{Timeout: 15 * time.Second}

// HandleProxyRequest routes requests by the first URL path segment and forwards them to target services.
func HandleProxyRequest(w http.ResponseWriter, r *http.Request) {
	recipientName, recipientPath, ok := extractRecipientName(r.URL.Path)
	if !ok {
		httphelpers.WriteError(w, http.StatusBadGateway, "Cannot process request")
		return
	}

	recipientURL := resolveRecipientURL(recipientName)
	if recipientURL == "" {
		httphelpers.WriteError(w, http.StatusBadGateway, "Cannot process request")
		return
	}

	targetURL, err := buildTargetURL(recipientURL, recipientPath, r.URL.RawQuery)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadGateway, "Cannot process request")
		return
	}

	forwardedRequest, err := http.NewRequestWithContext(r.Context(), r.Method, targetURL, r.Body)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadGateway, "Cannot process request")
		return
	}

	copyHeaders(forwardedRequest.Header, r.Header)

	response, err := proxyHTTPClient.Do(forwardedRequest)
	if err != nil {
		httphelpers.WriteError(w, http.StatusBadGateway, "Cannot process request")
		return
	}
	defer response.Body.Close()

	copyHeaders(w.Header(), response.Header)
	w.WriteHeader(response.StatusCode)
	_, _ = io.Copy(w, response.Body)
}

func extractRecipientName(path string) (recipientName string, recipientPath string, ok bool) {
	trimmedPath := strings.TrimPrefix(path, "/")
	if trimmedPath == "" {
		return "", "", false
	}

	segments := strings.Split(trimmedPath, "/")
	recipientName = strings.TrimSpace(segments[0])
	if recipientName == "" {
		return "", "", false
	}

	if len(segments) > 1 {
		recipientPath = "/" + strings.Join(segments[1:], "/")
	}

	return recipientName, recipientPath, true
}

func resolveRecipientURL(recipientName string) string {
	switch strings.ToLower(strings.TrimSpace(recipientName)) {
	case "product":
		return strings.TrimSpace(os.Getenv("PRODUCT_URL"))
	case "cart":
		return strings.TrimSpace(os.Getenv("CART_URL"))
	default:
		return ""
	}
}

func buildTargetURL(recipientBaseURL, recipientPath, rawQuery string) (string, error) {
	baseURL, err := url.Parse(recipientBaseURL)
	if err != nil {
		return "", err
	}

	if recipientPath != "" {
		if strings.HasSuffix(baseURL.Path, "/") {
			baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + recipientPath
		} else {
			baseURL.Path = strings.TrimSuffix(baseURL.Path, "/") + recipientPath
		}
	}

	baseURL.RawQuery = rawQuery
	return baseURL.String(), nil
}

func copyHeaders(destination, source http.Header) {
	for key, values := range source {
		for _, value := range values {
			destination.Add(key, value)
		}
	}
}
