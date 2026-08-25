package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DnssecRecordsDataSourceModel struct {
	Domain  types.String                                                               `tfsdk:"domain"`
	Records supertypes.SetNestedObjectValueOf[DnssecRecordsDataSourceModelRecordsItem] `tfsdk:"records"`
}

func (m *DnssecRecordsDataSourceModel) Fill(ctx context.Context, items []apiclient.GetDnssecRecords200JSONResponseBody_Records) (diags diag.Diagnostics) {
	m.Records = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(items, func(item apiclient.GetDnssecRecords200JSONResponseBody_Records, _ int) DnssecRecordsDataSourceModelRecordsItem {
		var mm DnssecRecordsDataSourceModelRecordsItem
		diags.Append(mm.Fill(ctx, item)...)
		return mm
	}))
	return
}

type DnssecRecordsDataSourceModelRecordsItem struct {
	KeyTag     types.String `tfsdk:"key_tag"`
	Alg        types.String `tfsdk:"alg"`
	Digest     types.String `tfsdk:"digest"`
	DigestType types.String `tfsdk:"digest_type"`
	PubKey     types.String `tfsdk:"pub_key"`
}

func (m *DnssecRecordsDataSourceModelRecordsItem) Fill(ctx context.Context, item apiclient.GetDnssecRecords200JSONResponseBody_Records) (diags diag.Diagnostics) {
	m.KeyTag = types.StringPointerValue(item.KeyTag)
	m.Alg = types.StringPointerValue(item.Alg)
	m.Digest = types.StringPointerValue(item.Digest)
	m.DigestType = types.StringPointerValue(item.DigestType)
	m.PubKey = types.StringPointerValue(item.PubKey)
	return
}

func NewDnssecRecordsDataSource() datasource.DataSource {
	return &DnssecRecordsDataSource{}
}

var _ datasource.DataSource = &DnssecRecordsDataSource{}

type DnssecRecordsDataSource struct {
	baseDataSource
}

func (d *DnssecRecordsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dnssec_records"
}

func (d *DnssecRecordsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves all DNSSEC records associated with the domain at the registry.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"records": schema.SetNestedAttribute{
				MarkdownDescription: "All DNSSEC records.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[DnssecRecordsDataSourceModelRecordsItem](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"key_tag": schema.StringAttribute{
							Computed: true,
						},
						"alg": schema.StringAttribute{
							Computed: true,
						},
						"digest_type": schema.StringAttribute{
							Computed: true,
						},
						"digest": schema.StringAttribute{
							Computed: true,
						},
						"pub_key": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *DnssecRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnssecRecordsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetDnssecRecordsWithResponse(
		ctx,
		data.Domain.ValueString(),
		nil,
	)
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if httpResp.StatusCode() != http.StatusOK || httpResp.JSON200 == nil || httpResp.JSON200.Status == nil || *httpResp.JSON200.Status != "SUCCESS" {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	}

	var records []apiclient.GetDnssecRecords200JSONResponseBody_Records
	if httpResp.JSON200.Records != nil {
		records = lo.Values(*httpResp.JSON200.Records)
	}

	resp.Diagnostics.Append(data.Fill(ctx, records)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
