package apiclient

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/logging"
)

func New(baseUrl, apiKey, secretKey string) (*ClientWithResponses, error) {
	transport := http.DefaultTransport

	// Logging
	transport = logging.NewLoggingHTTPTransport(transport)

	// Retry
	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = &http.Client{Transport: transport}
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryClient.Logger = nil
	retryClient.RetryMax = 10
	transport = retryClient.StandardClient().Transport

	return NewClientWithResponses(
		baseUrl,
		WithHTTPClient(&http.Client{Transport: transport}),
		WithRequestEditorFn(func(ctx context.Context, req *http.Request) error {
			req.Header.Set("X-API-Key", apiKey)
			req.Header.Set("X-Secret-API-Key", secretKey)
			return nil
		}),
	)
}

func IsOK[T any](response interface {
	StatusCode() int
	GetJSON200() *T
	GetBody() []byte
}) bool {
	if response.StatusCode() != http.StatusOK {
		return false
	}
	if response.GetJSON200() == nil {
		return false
	}

	var data struct {
		Status string `json:"status"`
	}
	if err := json.Unmarshal(response.GetBody(), &data); err != nil {
		return false
	}
	if data.Status != "SUCCESS" {
		return false
	}

	return true
}
