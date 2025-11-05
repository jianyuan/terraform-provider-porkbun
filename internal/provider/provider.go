package provider

import (
	"context"
	"os"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/provider/provider_porkbun"
)

// Ensure PorkbunProvider satisfies various provider interfaces.
var _ provider.Provider = &PorkbunProvider{}
var _ provider.ProviderWithFunctions = &PorkbunProvider{}

// PorkbunProvider defines the provider implementation.
type PorkbunProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

func (p *PorkbunProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "porkbun"
	resp.Version = p.version
}

func (p *PorkbunProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = provider_porkbun.PorkbunProviderSchema(ctx)
}

func (p *PorkbunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data provider_porkbun.PorkbunModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var baseUrl string
	if !data.BaseUrl.IsNull() {
		baseUrl = data.BaseUrl.ValueString()
	} else if v := os.Getenv("PORKBUN_BASE_URL"); v != "" {
		baseUrl = v
	} else {
		baseUrl = "https://api.porkbun.com/api/json"
	}

	var apiKey string
	if !data.ApiKey.IsNull() {
		apiKey = data.ApiKey.ValueString()
	} else if v := os.Getenv("PORKBUN_API_KEY"); v != "" {
		apiKey = v
	}

	var secretKey string
	if !data.SecretKey.IsNull() {
		secretKey = data.SecretKey.ValueString()
	} else if v := os.Getenv("PORKBUN_SECRET_KEY"); v != "" {
		secretKey = v
	}

	if baseUrl == "" {
		resp.Diagnostics.AddError("base_url is required", "base_url is required")
		return
	}

	if apiKey == "" {
		resp.Diagnostics.AddError("api_key is required", "api_key is required")
		return
	}

	if secretKey == "" {
		resp.Diagnostics.AddError("secret_key is required", "secret_key is required")
		return
	}

	retryClient := retryablehttp.NewClient()
	retryClient.ErrorHandler = retryablehttp.PassthroughErrorHandler
	retryClient.Logger = nil
	retryClient.RetryMax = 10

	client, err := apiclient.NewClientWithResponses(
		baseUrl,
		apiclient.WithHTTPClient(retryClient.StandardClient()),
	)
	if err != nil {
		resp.Diagnostics.AddError("failed to create API client", err.Error())
		return
	}

	providerData := &ProviderData{
		client:    client,
		apiKey:    apiKey,
		secretKey: secretKey,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PorkbunProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsRecordResource,
		NewDomainNameserversResource,
	}
}

func (p *PorkbunProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDnsRecordDataSource,
		NewDnsRecordsDataSource,
		NewDomainNameserversDataSource,
		NewDomainsDataSource,
	}
}

func (p *PorkbunProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &PorkbunProvider{
			version: version,
		}
	}
}

type ProviderData struct {
	client    *apiclient.ClientWithResponses
	apiKey    string
	secretKey string
}
