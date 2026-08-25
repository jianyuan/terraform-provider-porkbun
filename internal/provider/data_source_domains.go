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
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DomainsDomainDataSourceModel struct {
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

func (m *DomainsDomainDataSourceModel) Fill(ctx context.Context, domain apiclient.DomainListAllResponse_Domains) (diags diag.Diagnostics) {
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

type DomainsDataSourceModel struct {
	Domains supertypes.SetNestedObjectValueOf[DomainsDomainDataSourceModel] `tfsdk:"domains"`
}

func (m *DomainsDataSourceModel) Fill(ctx context.Context, domains []apiclient.DomainListAllResponse_Domains) (diags diag.Diagnostics) {
	m.Domains = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(domains, func(item apiclient.DomainListAllResponse_Domains, _ int) DomainsDomainDataSourceModel {
		var mm DomainsDomainDataSourceModel
		diags.Append(mm.Fill(ctx, item)...)
		return mm
	}))
	return
}

func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

var _ datasource.DataSource = &DomainsDataSource{}
var _ datasource.DataSourceWithConfigure = &DomainsDataSource{}

type DomainsDataSource struct {
	baseDataSource
}

func (d *DomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *DomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get all domain names in account.",

		Attributes: map[string]schema.Attribute{
			"domains": schema.SetNestedAttribute{
				MarkdownDescription: "Domain names in account.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[DomainsDomainDataSourceModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							Computed: true,
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
				},
			},
		},
	}
}

func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &apiclient.GetDomainsParams{
		Start: new(int64(0)),
	}

	var domains []apiclient.DomainListAllResponse_Domains

	for {
		httpResp, err := d.client.GetDomainsWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
			return
		} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
			resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
			return
		}

		domains = append(domains, httpResp.JSON200.Domains...)

		if len(httpResp.JSON200.Domains) == 0 {
			break
		}

		params.Start = new(int64(len(domains) + 1))
	}

	resp.Diagnostics.Append(data.Fill(ctx, domains)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
