package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

type DnsRecordDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`
	DnsRecordModel
}

func (m *DnsRecordDataSourceModel) FromAPI(ctx context.Context, record apiclient.DnsRecordsResponse_Records) (diags diag.Diagnostics) {
	diags.Append(m.DnsRecordModel.FromAPI(ctx, record)...)
	return
}

func NewDnsRecordDataSource() datasource.DataSource {
	return &DnsRecordDataSource{}
}

var _ datasource.DataSource = &DnsRecordDataSource{}

type DnsRecordDataSource struct {
	baseDataSource
}

func (d *DnsRecordDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (d *DnsRecordDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
		MarkdownDescription: "Retrieve a single record for a particular record ID.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"id": schema.StringAttribute{
				MarkdownDescription: "The record ID.",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The full name of the record being created, including the subdomain and the domain itself.",
				Computed:            true,
			},
			"type": schema.StringAttribute{
				MarkdownDescription: "The DNS record type.",
				Computed:            true,
			},
			"content": schema.StringAttribute{
				MarkdownDescription: "The record value.",
				Computed:            true,
			},
			"ttl": schema.Int64Attribute{
				MarkdownDescription: "Time to live in seconds. Minimum is determined by account settings (typically 600). Defaults to the account minimum if omitted or 0.",
				Computed:            true,
			},
			"priority": schema.Int64Attribute{
				MarkdownDescription: "The priority for `MX` and `SRV` records. Defaults to 0 if omitted.",
				Computed:            true,
			},
			"notes": schema.StringAttribute{
				MarkdownDescription: "Notes for the DNS record. Not served in DNS.",
				Computed:            true,
			},
		},
	}, fwschema.ForceNew(), fwschema.Enum())
}

func (d *DnsRecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnsRecordDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetDnsRecordByIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		nil,
	)
	httpResp.GetJSON200()
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	} else if len(httpResp.JSON200.Records) != 1 {
		resp.Diagnostics.AddError("Client error", fmt.Sprintf("Expected exactly one record, got %d", len(httpResp.JSON200.Records)))
		return
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, httpResp.JSON200.Records[0])...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
