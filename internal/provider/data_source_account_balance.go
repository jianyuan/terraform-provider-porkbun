package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
)

type AccountBalanceDataSourceModel struct {
	Balance types.Int64  `tfsdk:"balance"`
	Display types.String `tfsdk:"display"`
}

func (m *AccountBalanceDataSourceModel) FromAPI(ctx context.Context, data apiclient.BalanceResponse) (diags diag.Diagnostics) {
	m.Balance = types.Int64PointerValue(data.Balance)
	m.Display = types.StringPointerValue(data.Display)
	return
}

func NewAccountBalanceDataSource() datasource.DataSource {
	return &AccountBalanceDataSource{}
}

var _ datasource.DataSource = &AccountBalanceDataSource{}
var _ datasource.DataSourceWithConfigure = &AccountBalanceDataSource{}

type AccountBalanceDataSource struct {
	baseDataSource
}

func (d *AccountBalanceDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_balance"
}

func (d *AccountBalanceDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Returns the available account credit balance.",
		Attributes: map[string]schema.Attribute{
			"balance": schema.Int64Attribute{
				Description: "Available account credit balance in cents.",
				Computed:    true,
			},
			"display": schema.StringAttribute{
				Description: "Human-readable balance string (e.g. `$12.34`).",
				Computed:    true,
			},
		},
	}
}

func (d *AccountBalanceDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data AccountBalanceDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetBalanceWithResponse(ctx, nil)
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
