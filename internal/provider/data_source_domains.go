package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DomainsDataSourceModel struct {
	Domains supertypes.SetNestedObjectValueOf[DomainModel] `tfsdk:"domains"`
}

func (m *DomainsDataSourceModel) FromAPI(ctx context.Context, domains []apiclient.Domain) (diags diag.Diagnostics) {
	m.Domains = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(domains, func(item apiclient.Domain, _ int) DomainModel {
		var mm DomainModel
		diags.Append(mm.FromAPI(ctx, item)...)
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
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
		MarkdownDescription: "Get all domain names in account.",
		Attributes: map[string]schema.Attribute{
			"domains": schema.SetNestedAttribute{
				MarkdownDescription: "Domain names in account.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[DomainModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					CustomType: supertypes.NewObjectTypeOf[DomainModel](ctx),
					Attributes: map[string]schema.Attribute{
						"domain": schema.StringAttribute{
							MarkdownDescription: "Fully qualified domain name in the authenticated account.",
							Computed:            true,
						},
						"status": schema.StringAttribute{
							MarkdownDescription: "The status of the domain.",
							Computed:            true,
						},
						"tld": schema.StringAttribute{
							MarkdownDescription: "The TLD of the domain.",
							Computed:            true,
						},
						"create_date": schema.StringAttribute{
							MarkdownDescription: "The creation date of the domain.",
							Computed:            true,
						},
						"expire_date": schema.StringAttribute{
							MarkdownDescription: "The expiration date of the domain.",
							Computed:            true,
						},
						"security_lock": schema.BoolAttribute{
							MarkdownDescription: "Indicates if the Registrar Lock/Transfer Lock is enabled to prevent unauthorized domain transfers.",
							Computed:            true,
						},
						"whois_privacy": schema.BoolAttribute{
							MarkdownDescription: "Indicates if WHOIS privacy protection is active to hide personal registrant details from the public WHOIS database.",
							Computed:            true,
						},
						"auto_renew": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether the domain will automatically renew before its expiration date.",
							Computed:            true,
						},
						"api_access": schema.BoolAttribute{
							MarkdownDescription: "Indicates whether API access is enabled for the domain.",
							Computed:            true,
						},
						"not_local": schema.BoolAttribute{
							MarkdownDescription: "Indicates if the domain is registered with an external, third-party registrar.",
							Computed:            true,
						},
					},
				},
			},
		},
	}, fwschema.ForceNew(), fwschema.Enum())
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

	var domains []apiclient.Domain

	for {
		httpResp, err := d.client.GetDomainsWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
			return
		} else if !apiclient.IsOK(httpResp) {
			resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
			return
		}

		domains = append(domains, httpResp.JSON200.Domains...)

		if len(httpResp.JSON200.Domains) == 0 {
			break
		}

		params.Start = new(int64(len(domains) + 1))
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, domains)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
