package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type GlueRecordsDataSourceModel struct {
	Domain types.String                                                          `tfsdk:"domain"`
	Hosts  supertypes.MapNestedObjectValueOf[GlueRecordsDataSourceModelHostItem] `tfsdk:"hosts"`
}

func (m *GlueRecordsDataSourceModel) FromAPI(ctx context.Context, glue []apiclient.GetDomainGlue200JSONResponseBody_Glue) (diags diag.Diagnostics) {
	outHosts := make(map[string]GlueRecordsDataSourceModelHostItem)

	for _, g := range glue {
		var mm GlueRecordsDataSourceModelHostItem
		ips := append(append([]string{}, g.V4...), g.V6...)
		diags.Append(mm.FromAPI(ctx, ips)...)
		outHosts[g.Host] = mm
	}

	m.Hosts = supertypes.NewMapNestedObjectValueOfValueMap(ctx, outHosts)
	return
}

type GlueRecordsDataSourceModelHostItem struct {
	Ips supertypes.SetValueOf[string] `tfsdk:"ips"`
}

func (m *GlueRecordsDataSourceModelHostItem) FromAPI(ctx context.Context, ips []string) (diags diag.Diagnostics) {
	m.Ips = supertypes.NewSetValueOfSlice(ctx, ips)
	return
}

func NewGlueRecordsDataSource() datasource.DataSource {
	return &GlueRecordsDataSource{}
}

var _ datasource.DataSource = &GlueRecordsDataSource{}
var _ datasource.DataSourceWithConfigure = &GlueRecordsDataSource{}

type GlueRecordsDataSource struct {
	baseDataSource
}

func (d *GlueRecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_glue_records"
}

func (d *GlueRecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve all glue records (host objects) registered under the domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain for which to retrieve glue records.",
				Required:            true,
			},
			"hosts": schema.MapNestedAttribute{
				MarkdownDescription: "Map of hostnames to their IP addresses.",
				Computed:            true,
				CustomType:          supertypes.NewMapNestedObjectTypeOf[GlueRecordsDataSourceModelHostItem](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"ips": schema.SetAttribute{
							MarkdownDescription: "The IP addresses (IPv4 and/or IPv6) associated with the host record.",
							Computed:            true,
							CustomType:          supertypes.NewSetTypeOf[string](ctx),
						},
					},
				},
			},
		},
	}
}

func (d *GlueRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data GlueRecordsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetDomainGlueWithResponse(
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
	} else if httpResp.JSON200.Glue == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no glue records")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200.Glue)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
