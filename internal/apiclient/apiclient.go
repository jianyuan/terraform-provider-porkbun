package apiclient

import (
	"context"
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
