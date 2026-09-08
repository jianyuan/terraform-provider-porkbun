package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/porkbuntypes"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
)

type APISpendSettingsDataSourceModel struct {
	MonthlySpend types.Int64                                                                   `tfsdk:"monthly_spend"`
	Settings     supertypes.SingleNestedObjectValueOf[APISpendSettingsDataSourceModelSettings] `tfsdk:"settings"`
}

func (m *APISpendSettingsDataSourceModel) FromAPI(ctx context.Context, data apiclient.ApiSettingsResponse) (diags diag.Diagnostics) {
	m.MonthlySpend = types.Int64PointerValue(data.MonthlySpend)
	m.Settings = func() supertypes.SingleNestedObjectValueOf[APISpendSettingsDataSourceModelSettings] {
		if data.Settings == nil {
			return supertypes.NewSingleNestedObjectValueOfNull[APISpendSettingsDataSourceModelSettings](ctx)
		}
		var mm APISpendSettingsDataSourceModelSettings
		diags.Append(mm.FromAPI(ctx, *data.Settings)...)
		return supertypes.NewSingleNestedObjectValueOf(ctx, &mm)
	}()
	return
}

type APISpendSettingsDataSourceModelSettings struct {
	AutoTopup         types.Bool  `tfsdk:"auto_topup"`
	LowBalanceAlert   types.Int64 `tfsdk:"low_balance_alert"`
	MonthlySpendLimit types.Int64 `tfsdk:"monthly_spend_limit"`
	TopupAmount       types.Int64 `tfsdk:"topup_amount"`
	TopupThreshold    types.Int64 `tfsdk:"topup_threshold"`
}

func (m *APISpendSettingsDataSourceModelSettings) FromAPI(ctx context.Context, data apiclient.ApiSettingsResponse_Settings) (diags diag.Diagnostics) {
	m.AutoTopup = types.BoolPointerValue(data.AutoTopup)
	m.LowBalanceAlert = porkbuntypes.NullableInt64Value(data.LowBalanceAlert)
	m.MonthlySpendLimit = porkbuntypes.NullableInt64Value(data.MonthlySpendLimit)
	m.TopupAmount = porkbuntypes.NullableInt64Value(data.TopupAmount)
	m.TopupThreshold = porkbuntypes.NullableInt64Value(data.TopupThreshold)
	return
}

func NewAPISpendSettingsDataSource() datasource.DataSource {
	return &APISpendSettingsDataSource{}
}

var _ datasource.DataSource = &APISpendSettingsDataSource{}
var _ datasource.DataSourceWithConfigure = &APISpendSettingsDataSource{}

type APISpendSettingsDataSource struct {
	baseDataSource
}

func (d *APISpendSettingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_api_spend_settings"
}

func (d *APISpendSettingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Returns the available account credit balance.",
		Attributes: map[string]schema.Attribute{
			"monthly_spend": schema.Int64Attribute{
				MarkdownDescription: "Total API spend in the current calendar month, in cents.",
				Computed:            true,
			},
			"settings": schema.SingleNestedAttribute{
				MarkdownDescription: "Human-readable balance string (e.g. `$12.34`).",
				Computed:            true,
				CustomType:          supertypes.NewSingleNestedObjectTypeOf[APISpendSettingsDataSourceModelSettings](ctx),
				Attributes: map[string]schema.Attribute{
					"auto_topup": schema.BoolAttribute{
						MarkdownDescription: "Whether automatic balance top-up is enabled.",
						Computed:            true,
					},
					"low_balance_alert": schema.Int64Attribute{
						MarkdownDescription: "Send an alert email when balance drops below this amount (cents). `null` = disabled.",
						Computed:            true,
					},
					"monthly_spend_limit": schema.Int64Attribute{
						MarkdownDescription: "Maximum API spend per calendar month in cents. `null` means no limit.",
						Computed:            true,
					},
					"topup_amount": schema.Int64Attribute{
						MarkdownDescription: "Amount to add during an auto top-up (cents).",
						Computed:            true,
					},
					"topup_threshold": schema.Int64Attribute{
						MarkdownDescription: "Trigger auto top-up when balance falls below this amount (cents). `null` = disabled.",
						Computed:            true,
					},
				},
			},
		},
	}
}

func (d *APISpendSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data APISpendSettingsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetApiSettingsWithResponse(ctx, nil)
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
