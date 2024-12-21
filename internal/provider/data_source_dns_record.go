package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type DnsRecordDataSourceModel struct {
	Domain types.String `tfsdk:"domain"`
	DnsRecordModel
}

func (m *DnsRecordDataSourceModel) Fill(ctx context.Context, record apiclient.DnsRecord) (diags diag.Diagnostics) {
	diags.Append(m.DnsRecordModel.Fill(ctx, record)...)
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
	resp.Schema = schema.Schema{
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
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"content": schema.StringAttribute{
				Computed: true,
			},
			"ttl": schema.Int64Attribute{
				Computed: true,
			},
			"priority": schema.Int64Attribute{
				Computed: true,
			},
			"notes": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *DnsRecordDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnsRecordDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.DnsRetrieveRecordsByDomainAndIdWithResponse(
		ctx,
		data.Domain.ValueString(),
		data.Id.ValueString(),
		apiclient.DnsRetrieveRecordsByDomainAndIdJSONRequestBody{
			Apikey:       d.apiKey,
			Secretapikey: d.secretKey,
		},
	)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got error: %s", err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read, got status code %d: %s", httpResp.StatusCode(), string(httpResp.Body)))
		return
	} else if len(httpResp.JSON200.Records) != 1 {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Expected exactly one record, got %d", len(httpResp.JSON200.Records)))
		return
	}

	resp.Diagnostics.Append(data.Fill(ctx, httpResp.JSON200.Records[0])...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
