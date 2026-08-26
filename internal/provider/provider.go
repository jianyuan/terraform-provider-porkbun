package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type PorkbunProviderModel struct {
	ApiKey    types.String `tfsdk:"api_key"`
	BaseUrl   types.String `tfsdk:"base_url"`
	SecretKey types.String `tfsdk:"secret_key"`
}

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
	resp.Schema = schema.Schema{
		MarkdownDescription: "The Porkbun provider is used to interact with the Porkbun service.\n\nIf you find this provider useful, please consider supporting me through GitHub Sponsorship or Ko-Fi to help with its development.\n\n[![Github-sponsors](https://img.shields.io/badge/sponsor-30363D?style=for-the-badge&logo=GitHub-Sponsors&logoColor=#EA4AAA)](https://github.com/sponsors/jianyuan)\n[![Ko-Fi](https://img.shields.io/badge/Ko--fi-F16061?style=for-the-badge&logo=ko-fi&logoColor=white)](https://ko-fi.com/L3L71DQEL)",
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "The API key for the Porkbun account. It can be sourced from the `PORKBUN_API_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "The base URL for the Porkbun API. Defaults to `https://api.porkbun.com/api/json/v3`. It can be sourced from the `PORKBUN_BASE_URL` environment variable.",
				Optional:            true,
			},
			"secret_key": schema.StringAttribute{
				MarkdownDescription: "The secret API key for the Porkbun account. It can be sourced from the `PORKBUN_SECRET_KEY` environment variable.",
				Optional:            true,
				Sensitive:           true,
			},
		},
	}
}

func (p *PorkbunProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data PorkbunProviderModel

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
		baseUrl = "https://api.porkbun.com/api/json/v3"
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

	client, err := apiclient.New(baseUrl, apiKey, secretKey)
	if err != nil {
		resp.Diagnostics.AddError("failed to create API client", err.Error())
		return
	}

	providerData := &ProviderData{
		client: client,
	}

	resp.DataSourceData = providerData
	resp.ResourceData = providerData
}

func (p *PorkbunProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewDnsRecordResource,
		NewDnssecRecordResource,
		NewDomainAutoRenewSettingResource,
		NewDomainNameserversResource,
	}
}

func (p *PorkbunProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewDnsRecordDataSource,
		NewDnsRecordsDataSource,
		NewDnssecRecordsDataSource,
		NewDomainDataSource,
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
	client *apiclient.ClientWithResponses
}
