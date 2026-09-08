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

func (m *GlueRecordsDataSourceModel) FromAPI(ctx context.Context, hosts [][]apiclient.GetDomainGlue200JSONResponseBody_Hosts) (diags diag.Diagnostics) {
	outHosts := make(map[string]GlueRecordsDataSourceModelHostItem)

	for _, host := range hosts {
		if len(host) != 2 {
			diags.AddError("Client error", "Unable to read, got unexpected glue record format")
			return
		}

		hostname, err := host[0].AsGetDomainGlue200JSONResponseBodyHosts0()
		if err != nil {
			diags.AddError("Client error", "Unable to read, got unexpected glue record format")
			return
		}

		var mm GlueRecordsDataSourceModelHostItem

		if ipAddresses, err := host[1].AsGetDomainGlue200JSONResponseBodyHosts1(); err == nil {
			diags.Append(mm.FromAPI(ctx, ipAddresses)...)
		} else if ipAddresses, err := host[1].AsGetDomainGlue200JSONResponseBodyHosts2(); err == nil {
			var ips []string
			if ipAddresses.V4 != nil {
				ips = append(ips, *ipAddresses.V4...)
			}
			if ipAddresses.V6 != nil {
				ips = append(ips, *ipAddresses.V6...)
			}
			diags.Append(mm.FromAPI(ctx, ips)...)
		} else {
			diags.AddError("Client error", "Unable to read, got unexpected glue record format")
			return
		}

		outHosts[hostname] = mm
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
	} else if httpResp.JSON200.Hosts == nil {
		resp.Diagnostics.AddError("Client error", "Unable to read, got no glue records")
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200.Hosts)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}
