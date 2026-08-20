package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
)

type DomainDataSourceModel struct {
	Domain       types.String `tfsdk:"domain"`
	Status       types.String `tfsdk:"status"`
	Tld          types.String `tfsdk:"tld"`
	CreateDate   types.String `tfsdk:"create_date"`
	ExpireDate   types.String `tfsdk:"expire_date"`
	SecurityLock types.Bool   `tfsdk:"security_lock"`
	WhoisPrivacy types.Bool   `tfsdk:"whois_privacy"`
	AutoRenew    types.Bool   `tfsdk:"auto_renew"`
	ApiAccess    types.Bool   `tfsdk:"api_access"`
	NotLocal     types.Bool   `tfsdk:"not_local"`
}

func (m *DomainDataSourceModel) Fill(ctx context.Context, domain apiclient.GetDomain200JSONResponseBody_Domain) (diags diag.Diagnostics) {
	m.Domain = types.StringPointerValue(domain.Domain)
	m.Status = types.StringPointerValue(domain.Status)
	m.Tld = types.StringPointerValue(domain.Tld)
	m.CreateDate = types.StringPointerValue(domain.CreateDate)
	m.ExpireDate = types.StringPointerValue(domain.ExpireDate)
	m.SecurityLock = porkbuntypes.FlexibleBoolPointerValue(domain.SecurityLock)
	m.WhoisPrivacy = porkbuntypes.FlexibleBoolPointerValue(domain.WhoisPrivacy)
	m.AutoRenew = porkbuntypes.FlexibleBoolPointerValue(domain.AutoRenew)
	m.ApiAccess = porkbuntypes.FlexibleBoolPointerValue(domain.ApiAccess)
	m.NotLocal = porkbuntypes.FlexibleBoolPointerValue(domain.NotLocal)
	return
}

func NewDomainDataSource() datasource.DataSource {
	return &DomainDataSource{}
}

var _ datasource.DataSource = &DomainDataSource{}
var _ datasource.DataSourceWithConfigure = &DomainDataSource{}

type DomainDataSource struct {
	baseDataSource
}

func (d *DomainDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (d *DomainDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the metadata for a single domain.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "Fully qualified domain name in the authenticated account.",
				Required:            true,
			},
			"status": schema.StringAttribute{
				Computed: true,
			},
			"tld": schema.StringAttribute{
				Computed: true,
			},
			"create_date": schema.StringAttribute{
				Computed: true,
			},
			"expire_date": schema.StringAttribute{
				Computed: true,
			},
			"security_lock": schema.BoolAttribute{
				Computed: true,
			},
			"whois_privacy": schema.BoolAttribute{
				Computed: true,
			},
			"auto_renew": schema.BoolAttribute{
				Computed: true,
			},
			"api_access": schema.BoolAttribute{
				Computed: true,
			},
			"not_local": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetDomainWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status == nil || *httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, *httpResp.JSON200.Domain)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
