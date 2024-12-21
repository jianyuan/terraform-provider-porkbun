package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/go-utils/ptr"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
	"github.com/jianyuan/terraform-provider-porkbun/internal/tfutils"
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
	NotLocal     types.Bool   `tfsdk:"not_local"`
	// Labels       types.Set    `tfsdk:"labels"`
}

func (m *DomainsDomainDataSourceModel) Fill(ctx context.Context, domain apiclient.Domain) (diags diag.Diagnostics) {
	m.Domain = types.StringValue(domain.Domain)
	m.Status = types.StringValue(domain.Status)
	m.Tld = types.StringValue(domain.Tld)
	m.CreateDate = types.StringValue(domain.CreateDate)
	m.ExpireDate = types.StringValue(domain.ExpireDate)
	m.SecurityLock = tfutils.MergeDiagnostics(porkbuntypes.StringIntegerBoolValue(domain.SecurityLock))(&diags)
	m.WhoisPrivacy = tfutils.MergeDiagnostics(porkbuntypes.StringIntegerBoolValue(domain.WhoisPrivacy))(&diags)
	m.AutoRenew = tfutils.MergeDiagnostics(porkbuntypes.StringIntegerBoolValue(domain.AutoRenew))(&diags)
	m.NotLocal = tfutils.MergeDiagnostics(porkbuntypes.IntegerBoolValue(domain.NotLocal))(&diags)
	// m.Labels = types.SetValueMust(types.StringType, sliceutils.Map(func(v string) attr.Value {
	// 	return types.StringValue(v)
	// }, domain.Labels))
	return
}

type DomainsDataSourceModel struct {
	Domains []DomainsDomainDataSourceModel `tfsdk:"domains"`
}

func (m *DomainsDataSourceModel) Fill(ctx context.Context, domains []apiclient.Domain) (diags diag.Diagnostics) {
	m.Domains = make([]DomainsDomainDataSourceModel, len(domains))
	for i, domain := range domains {
		diags = append(diags, m.Domains[i].Fill(ctx, domain)...)
		if diags.HasError() {
			return
		}
	}
	return
}

func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

var _ datasource.DataSource = &DomainsDataSource{}

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
						"not_local": schema.BoolAttribute{
							Computed: true,
						},
						// "labels": schema.SetAttribute{
						// 	Computed:    true,
						// },
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

	params := apiclient.DomainListAllJSONRequestBody{
		Apikey:       d.apiKey,
		Secretapikey: d.secretKey,
		Start:        ptr.Ptr(0),
	}

	var domains []apiclient.Domain

	for {
		httpResp, err := d.client.DomainListAllWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
			return
		} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
			return
		}

		domains = append(domains, httpResp.JSON200.Domains...)

		if len(httpResp.JSON200.Domains) == 0 {
			break
		}

		params.Start = ptr.Ptr(len(domains) + 1)
	}

	resp.Diagnostics.Append(data.Fill(ctx, domains)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
