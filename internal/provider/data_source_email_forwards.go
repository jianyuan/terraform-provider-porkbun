package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-plugin-framework-utils/fwschema"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type EmailForwardsDataSourceModel struct {
	Domain        types.String                                                  `tfsdk:"domain"`
	DnsConfigured types.Bool                                                    `tfsdk:"dns_configured"`
	Limits        supertypes.SingleNestedObjectValueOf[EmailForwardLimitsModel] `tfsdk:"limits"`
	EmailForwards supertypes.SetNestedObjectValueOf[EmailForwardModel]          `tfsdk:"email_forwards"`
}

func (m *EmailForwardsDataSourceModel) FromAPI(ctx context.Context, data apiclient.GetEmailForwards200JSONResponseBody) (diags diag.Diagnostics) {
	m.DnsConfigured = types.BoolPointerValue(data.DnsConfigured)

	if data.Limits == nil {
		m.Limits = supertypes.NewSingleNestedObjectValueOfNull[EmailForwardLimitsModel](ctx)
	} else {
		var mm EmailForwardLimitsModel
		diags.Append(mm.FromAPI(ctx, *data.Limits)...)
		m.Limits = supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	}

	var forwards []apiclient.GetEmailForwards200JSONResponseBody_Forwards
	if data.Forwards != nil {
		forwards = *data.Forwards
	}
	m.EmailForwards = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(forwards, func(forward apiclient.GetEmailForwards200JSONResponseBody_Forwards, _ int) EmailForwardModel {
		var mm EmailForwardModel
		diags.Append(mm.FromAPI(ctx, forward)...)
		return mm
	}))
	return
}

func NewEmailForwardsDataSource() datasource.DataSource {
	return &EmailForwardsDataSource{}
}

var _ datasource.DataSource = &EmailForwardsDataSource{}

type EmailForwardsDataSource struct {
	baseDataSource
}

func (d *EmailForwardsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_forwards"
}

func (d *EmailForwardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = fwschema.Describe(ctx, schema.Schema{
		MarkdownDescription: "Retrieve all email forwards associated with a domain.",
		Attributes: map[string]schema.Attribute{
			"domain": schema.StringAttribute{
				MarkdownDescription: "The domain name.",
				Required:            true,
			},
			"dns_configured": schema.BoolAttribute{
				MarkdownDescription: "Whether the domain's DNS is configured for email forwarding.",
				Computed:            true,
			},
			"limits": schema.SingleNestedAttribute{
				MarkdownDescription: "The forwards the domain has (`used`, including any Porkbun manages itself) and may have (`max`).",
				Computed:            true,
				CustomType:          supertypes.NewSingleNestedObjectTypeOf[EmailForwardLimitsModel](ctx),
				Attributes: map[string]schema.Attribute{
					"used": schema.Int64Attribute{
						MarkdownDescription: "The number of forwards in use.",
						Computed:            true,
					},
					"max": schema.Int64Attribute{
						MarkdownDescription: "The maximum number of forwards allowed.",
						Computed:            true,
					},
				},
			},
			"email_forwards": schema.SetNestedAttribute{
				MarkdownDescription: "List of email forwards.",
				Computed:            true,
				CustomType:          supertypes.NewSetNestedObjectTypeOf[EmailForwardModel](ctx),
				NestedObject: schema.NestedAttributeObject{
					CustomType: supertypes.NewObjectTypeOf[EmailForwardModel](ctx),
					Attributes: map[string]schema.Attribute{
						"address": schema.StringAttribute{
							MarkdownDescription: "The forwarded address: the part before the `@`.",
							Computed:            true,
						},
						"forward_to": schema.StringAttribute{
							MarkdownDescription: "The mailbox mail is delivered to.",
							Computed:            true,
						},
						"created": schema.StringAttribute{
							MarkdownDescription: "When the forward was created.",
							Computed:            true,
						},
					},
				},
			},
		},
	}, fwschema.DefaultSteps()...)
}

func (d *EmailForwardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data EmailForwardsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetEmailForwardsWithResponse(ctx, data.Domain.ValueString(), apiclient.GetEmailForwardsJSONRequestBody{})
	if err != nil {
		resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
		return
	} else if !apiclient.IsOK(httpResp) {
		resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
		return
	}
	resp.Diagnostics.Append(data.FromAPI(ctx, *httpResp.JSON200)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
