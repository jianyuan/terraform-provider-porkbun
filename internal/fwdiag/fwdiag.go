package fwdiag

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	intfwdiag "github.com/jianyuan/terraform-plugin-framework-utils/fwdiag"
)

func Merge[T any](v T, sourceDiags diag.Diagnostics) func(targetDiags *diag.Diagnostics) T {
	return intfwdiag.Merge(v, sourceDiags)
}

func NewClientError(action string, err error) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Client error", fmt.Sprintf("Unable to %s, got error: %s", action, err))
}

func NewClientReadError(err error) diag.ErrorDiagnostic {
	return NewClientError("read", err)
}

func NewClientCreateError(err error) diag.ErrorDiagnostic {
	return NewClientError("create", err)
}

func NewClientUpdateError(err error) diag.ErrorDiagnostic {
	return NewClientError("update", err)
}

func NewClientDeleteError(err error) diag.ErrorDiagnostic {
	return NewClientError("delete", err)
}

type HTTPResponse interface {
	StatusCode() int
	GetBody() []byte
}

func NewClientHTTPResponseError(action string, httpResp HTTPResponse) diag.ErrorDiagnostic {
	return diag.NewErrorDiagnostic("Client error", fmt.Sprintf("Unable to %s, got status code %d: %s", action, httpResp.StatusCode(), string(httpResp.GetBody())))
}

func NewClientReadHTTPResponseError(httpResp HTTPResponse) diag.ErrorDiagnostic {
	return NewClientHTTPResponseError("read", httpResp)
}

func NewClientCreateHTTPResponseError(httpResp HTTPResponse) diag.ErrorDiagnostic {
	return NewClientHTTPResponseError("create", httpResp)
}

func NewClientUpdateHTTPResponseError(httpResp HTTPResponse) diag.ErrorDiagnostic {
	return NewClientHTTPResponseError("update", httpResp)
}

func NewClientDeleteHTTPResponseError(httpResp HTTPResponse) diag.ErrorDiagnostic {
	return NewClientHTTPResponseError("delete", httpResp)
}
