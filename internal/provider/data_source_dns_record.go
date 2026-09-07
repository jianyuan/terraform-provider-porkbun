package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
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
	resp.Schema = dnsRecordSchema().GetDataSource(ctx)
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
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
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
