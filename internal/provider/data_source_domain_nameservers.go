package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/go-utils/sliceutils"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type DomainNameserversDataSourceModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.Set    `tfsdk:"nameservers"`
}

func (m *DomainNameserversDataSourceModel) Fill(ctx context.Context, nameservers []string) (diags diag.Diagnostics) {
	m.Nameservers = types.SetValueMust(types.StringType, sliceutils.Map(func(v string) attr.Value {
		return types.StringValue(v)
	}, nameservers))
	return
}

func NewDomainNameserversDataSource() datasource.DataSource {
	return &DomainNameserversDataSource{}
}

var _ datasource.DataSource = &DomainNameserversDataSource{}

type DomainNameserversDataSource struct {
	baseDataSource
}

func (d *DomainNameserversDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain_nameservers"
}

func (d *DomainNameserversDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Get the authoritative name servers listed at the registry for your domain.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name for which to retrieve the authoritative name servers.",
				Required:            true,
			},
			"nameservers": schema.SetAttribute{
				Computed:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (d *DomainNameserversDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainNameserversDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.DomainGetNameServersWithResponse(
		ctx,
		data.Domain.ValueString(),
		apiclient.DomainGetNameServersJSONRequestBody{
			Apikey:       d.apiKey,
			Secretapikey: d.secretKey,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, httpResp.JSON200.Ns)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
