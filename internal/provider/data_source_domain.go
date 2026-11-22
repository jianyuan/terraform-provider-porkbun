package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

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
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
		MarkdownDescription: "Get the metadata for a single domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "Fully qualified domain name in the authenticated account.",
				Required:            true,
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
	}, fwschema.DefaultSteps()...)
}

func (d *DomainDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainModel

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
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200.Domain)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
