package provider

import (
	"context"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type DnsRecordsFilterDataSourceModel struct {
	Type      types.String `tfsdk:"type"`
	Subdomain types.String `tfsdk:"subdomain"`
}

type DnsRecordsDataSourceModel struct {
	Domain  types.String                     `tfsdk:"domain"`
	Filter  *DnsRecordsFilterDataSourceModel `tfsdk:"filter"`
	Records []DnsRecordModel                 `tfsdk:"records"`
}

func (m *DnsRecordsDataSourceModel) Fill(ctx context.Context, records []apiclient.DnsRecord) (diags diag.Diagnostics) {
	m.Records = make([]DnsRecordModel, len(records))
	for i, record := range records {
		diags = append(diags, m.Records[i].Fill(ctx, record)...)
		if diags.HasError() {
			return
		}
	}
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

	var records []apiclient.DnsRecord
	if data.Filter != nil && !data.Filter.Type.IsNull() {
		if !data.Filter.Subdomain.IsNull() {
			httpResp, err := d.client.DnsRetrieveRecordsByDomainAndTypeAndSubdomainWithResponse(
				ctx,
				data.Domain.ValueString(),
				data.Filter.Type.ValueString(),
				data.Filter.Subdomain.ValueString(),
				apiclient.DnsRetrieveRecordsByDomainAndTypeAndSubdomainJSONRequestBody{
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
			}

			records = httpResp.JSON200.Records
		} else {
			httpResp, err := d.client.DnsRetrieveRecordsByDomainAndTypeWithResponse(
				ctx,
				data.Domain.ValueString(),
				data.Filter.Type.ValueString(),
				apiclient.DnsRetrieveRecordsByDomainAndTypeJSONRequestBody{
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
			}

			records = httpResp.JSON200.Records
		}
	} else {
		httpResp, err := d.client.DnsRetrieveRecordsByDomainWithResponse(
			ctx,
			data.Domain.ValueString(),
			apiclient.DnsRetrieveRecordsByDomainJSONRequestBody{
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
		}

		records = httpResp.JSON200.Records
	}

	resp.Diagnostics.Append(data.Fill(ctx, records)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
