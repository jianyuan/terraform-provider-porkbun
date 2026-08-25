package provider

import (
	"context"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/tfutils"
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

func (m *DnsRecordsDataSourceModel) Fill(ctx context.Context, records []apiclient.DnsRecordsResponse_Records) (diags diag.Diagnostics) {
	m.Records = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(records, func(record apiclient.DnsRecordsResponse_Records, _ int) DnsRecordModel {
		var mm DnsRecordModel
		diags.Append(mm.Fill(ctx, record)...)
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
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieve all editable DNS records associated with a domain.",

		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"filter": schema.SingleNestedAttribute{
				MarkdownDescription: "Record filter.",
				Optional:            true,
				CustomType:          supertypes.NewSingleNestedObjectTypeOf[DnsRecordsFilterDataSourceModel](ctx),
				Attributes: map[string]schema.Attribute{
					"type": schema.StringAttribute{
						MarkdownDescription: "Record type. Valid types are: A, MX, CNAME, ALIAS, TXT, NS, AAAA, SRV, TLSA, CAA, HTTPS, SVCB.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.OneOf(DnsRecordTypes...),
						},
					},
					"subdomain": schema.StringAttribute{
						MarkdownDescription: "Record subdomain.",
						Optional:            true,
						Validators: []validator.String{
							stringvalidator.AlsoRequires(path.MatchRelative().AtParent().AtName("type")),
						},
					},
				},
			},
			"records": schema.SetNestedAttribute{
				MarkdownDescription: "All editable DNS records.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[DnsRecordModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
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
				},
			},
		},
	}
}

func (d *DnsRecordsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DnsRecordsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var records []apiclient.DnsRecordsResponse_Records
	if data.Filter.IsKnown() {
		filter := tfutils.MergeDiagnostics(data.Filter.Get(ctx))(&resp.Diagnostics)
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

	resp.Diagnostics.Append(data.Fill(ctx, records)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
