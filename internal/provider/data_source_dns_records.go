package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
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
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
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
						MarkdownDescription: "Record type.",
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
					CustomType: supertypes.NewObjectTypeOf[DnsRecordModel](ctx),
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							MarkdownDescription: "The record ID.",
							Computed:            true,
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
				},
			},
		},
	}, fwschema.DefaultSteps()...)
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
		} else if !apiclient.IsOK(httpResp) {
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
		} else if !apiclient.IsOK(httpResp) {
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
