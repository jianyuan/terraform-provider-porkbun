package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type URLForwardsDataSourceModel struct {
	Domain      types.String                                       `tfsdk:"domain"`
	UrlForwards supertypes.SetNestedObjectValueOf[URLForwardModel] `tfsdk:"url_forwards"`
}

func (m *URLForwardsDataSourceModel) FromAPI(ctx context.Context, forwards []apiclient.GetUrlForwardingResponse_Forwards) (diags diag.Diagnostics) {
	m.UrlForwards = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(forwards, func(forward apiclient.GetUrlForwardingResponse_Forwards, _ int) URLForwardModel {
		var mm URLForwardModel
		diags.Append(mm.FromAPI(ctx, forward)...)
		return mm
	}))
	return
}

func NewURLForwardsDataSource() datasource.DataSource {
	return &URLForwardsDataSource{}
}

var _ datasource.DataSource = &URLForwardsDataSource{}

type URLForwardsDataSource struct {
	baseDataSource
}

func (d *URLForwardsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_url_forwards"
}

func (d *URLForwardsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = urlForwardsSchema().GetDataSource(ctx)
}

func (d *URLForwardsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data URLForwardsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	httpResp, err := d.client.GetDomainUrlForwardingWithResponse(
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

	resp.Diagnostics.Append(data.FromAPI(ctx, httpResp.JSON200.Forwards)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
