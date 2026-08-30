package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DnsRecordsFilterDataSourceModel struct {
	Type      types.String `tfsdk:"type"`
	Subdomain types.String `tfsdk:"subdomain"`
}

type DnsRecordsDataSourceModel struct {
	Domain  types.String                                                          `tfsdk:"domain"`
	Filter  supertypes.SingleNestedObjectValueOf[DnsRecordsFilterDataSourceModel] `tfsdk:"filter"`
	Records supertypes.SetNestedObjectValueOf[DnsRecordModel]                     `tfsdk:"records"`
}

func (m *DnsRecordsDataSourceModel) FromAPI(ctx context.Context, records []apiclient.DnsRecordsResponse_Records) (diags diag.Diagnostics) {
	m.Records = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(records, func(record apiclient.DnsRecordsResponse_Records, _ int) DnsRecordModel {
		var mm DnsRecordModel
		diags.Append(mm.FromAPI(ctx, record)...)
		return mm
	}))
	return
}

func NewDnsRecordsDataSource() datasource.DataSource {
	return &DnsRecordsDataSource{}
}

var _ datasource.DataSource = &DnsRecordsDataSource{}

type DnsRecordsDataSource struct {
	baseDataSource
}

func (d *DnsRecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_records"
}

func (d *DnsRecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = dnsRecordsSchema().GetDataSource(ctx)
}

func (d *DnsRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnsRecordsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var records []apiclient.DnsRecordsResponse_Records
	if data.Filter.IsKnown() {
		filter := fwdiag.Merge(data.Filter.Get(ctx))(&resp.Diagnostics)
		if resp.Diagnostics.HasError() {
			return
		}

		httpResp, err := d.client.GetDnsRecordsByNameTypeWithResponse(
			ctx,
			data.Domain.ValueString(),
			filter.Type.ValueString(),
			filter.Subdomain.ValueString(),
			nil,
		)
		if err != nil {
			resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
			return
		} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
			resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
			return
		}

		records = httpResp.JSON200.Records

	} else {
		httpResp, err := d.client.GetDnsRecordsWithResponse(
			ctx,
			data.Domain.ValueString(),
			nil,
		)
		if err != nil {
			resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
			return
		} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status != "SUCCESS" {
			resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
			return
		}

		records = httpResp.JSON200.Records
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, records)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
