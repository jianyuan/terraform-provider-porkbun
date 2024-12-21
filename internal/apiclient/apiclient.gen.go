// Package apiclient provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.1 DO NOT EDIT.
package apiclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

// Defines values for DomainListAllJSONBodyIncludeLabels.
const (
	Yes DomainListAllJSONBodyIncludeLabels = "yes"
)

// Domain defines model for Domain.
type Domain struct {
	AutoRenew  string `json:"autoRenew"`
	CreateDate string `json:"createDate"`
	Domain     string `json:"domain"`
	ExpireDate string `json:"expireDate"`
	Labels     []struct {
		Color string `json:"color"`
		Id    string `json:"id"`
		Title string `json:"title"`
	} `json:"labels"`
	NotLocal     int    `json:"notLocal"`
	SecurityLock string `json:"securityLock"`
	Status       string `json:"status"`
	Tld          string `json:"tld"`
	WhoisPrivacy string `json:"whoisPrivacy"`
}

// DomainGetNameServersResponse defines model for DomainGetNameServersResponse.
type DomainGetNameServersResponse struct {
	Ns     []string `json:"ns"`
	Status string   `json:"status"`
}

// DomainListAllResponse defines model for DomainListAllResponse.
type DomainListAllResponse struct {
	Domains []Domain `json:"domains"`
	Status  string   `json:"status"`
}

// DomainGetNameServersJSONBody defines parameters for DomainGetNameServers.
type DomainGetNameServersJSONBody struct {
	Apikey       string `json:"apikey"`
	Secretapikey string `json:"secretapikey"`
}

// DomainListAllJSONBody defines parameters for DomainListAll.
type DomainListAllJSONBody struct {
	Apikey        string                              `json:"apikey"`
	IncludeLabels *DomainListAllJSONBodyIncludeLabels `json:"includeLabels,omitempty"`
	Secretapikey  string                              `json:"secretapikey"`
	Start         *int                                `json:"start,omitempty"`
}

// DomainListAllJSONBodyIncludeLabels defines parameters for DomainListAll.
type DomainListAllJSONBodyIncludeLabels string

// DomainGetNameServersJSONRequestBody defines body for DomainGetNameServers for application/json ContentType.
type DomainGetNameServersJSONRequestBody DomainGetNameServersJSONBody

// DomainListAllJSONRequestBody defines body for DomainListAll for application/json ContentType.
type DomainListAllJSONRequestBody DomainListAllJSONBody

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// DomainGetNameServersWithBody request with any body
	DomainGetNameServersWithBody(ctx context.Context, domain string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	DomainGetNameServers(ctx context.Context, domain string, body DomainGetNameServersJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// DomainListAllWithBody request with any body
	DomainListAllWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	DomainListAll(ctx context.Context, body DomainListAllJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) DomainGetNameServersWithBody(ctx context.Context, domain string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDomainGetNameServersRequestWithBody(c.Server, domain, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DomainGetNameServers(ctx context.Context, domain string, body DomainGetNameServersJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDomainGetNameServersRequest(c.Server, domain, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DomainListAllWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDomainListAllRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) DomainListAll(ctx context.Context, body DomainListAllJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewDomainListAllRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewDomainGetNameServersRequest calls the generic DomainGetNameServers builder with application/json body
func NewDomainGetNameServersRequest(server string, domain string, body DomainGetNameServersJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewDomainGetNameServersRequestWithBody(server, domain, "application/json", bodyReader)
}

// NewDomainGetNameServersRequestWithBody generates requests for DomainGetNameServers with any type of body
func NewDomainGetNameServersRequestWithBody(server string, domain string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "domain", runtime.ParamLocationPath, domain)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/v3/domain/getNs/%s", pathParam0)
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewDomainListAllRequest calls the generic DomainListAll builder with application/json body
func NewDomainListAllRequest(server string, body DomainListAllJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewDomainListAllRequestWithBody(server, "application/json", bodyReader)
}

// NewDomainListAllRequestWithBody generates requests for DomainListAll with any type of body
func NewDomainListAllRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/v3/domain/listAll")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// DomainGetNameServersWithBodyWithResponse request with any body
	DomainGetNameServersWithBodyWithResponse(ctx context.Context, domain string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DomainGetNameServersResp, error)

	DomainGetNameServersWithResponse(ctx context.Context, domain string, body DomainGetNameServersJSONRequestBody, reqEditors ...RequestEditorFn) (*DomainGetNameServersResp, error)

	// DomainListAllWithBodyWithResponse request with any body
	DomainListAllWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DomainListAllResp, error)

	DomainListAllWithResponse(ctx context.Context, body DomainListAllJSONRequestBody, reqEditors ...RequestEditorFn) (*DomainListAllResp, error)
}

type DomainGetNameServersResp struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DomainGetNameServersResponse
}

// Status returns HTTPResponse.Status
func (r DomainGetNameServersResp) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DomainGetNameServersResp) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type DomainListAllResp struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *DomainListAllResponse
}

// Status returns HTTPResponse.Status
func (r DomainListAllResp) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r DomainListAllResp) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// DomainGetNameServersWithBodyWithResponse request with arbitrary body returning *DomainGetNameServersResp
func (c *ClientWithResponses) DomainGetNameServersWithBodyWithResponse(ctx context.Context, domain string, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DomainGetNameServersResp, error) {
	rsp, err := c.DomainGetNameServersWithBody(ctx, domain, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDomainGetNameServersResp(rsp)
}

func (c *ClientWithResponses) DomainGetNameServersWithResponse(ctx context.Context, domain string, body DomainGetNameServersJSONRequestBody, reqEditors ...RequestEditorFn) (*DomainGetNameServersResp, error) {
	rsp, err := c.DomainGetNameServers(ctx, domain, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDomainGetNameServersResp(rsp)
}

// DomainListAllWithBodyWithResponse request with arbitrary body returning *DomainListAllResp
func (c *ClientWithResponses) DomainListAllWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*DomainListAllResp, error) {
	rsp, err := c.DomainListAllWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDomainListAllResp(rsp)
}

func (c *ClientWithResponses) DomainListAllWithResponse(ctx context.Context, body DomainListAllJSONRequestBody, reqEditors ...RequestEditorFn) (*DomainListAllResp, error) {
	rsp, err := c.DomainListAll(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseDomainListAllResp(rsp)
}

// ParseDomainGetNameServersResp parses an HTTP response from a DomainGetNameServersWithResponse call
func ParseDomainGetNameServersResp(rsp *http.Response) (*DomainGetNameServersResp, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DomainGetNameServersResp{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DomainGetNameServersResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}

// ParseDomainListAllResp parses an HTTP response from a DomainListAllWithResponse call
func ParseDomainListAllResp(rsp *http.Response) (*DomainListAllResp, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &DomainListAllResp{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest DomainListAllResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}