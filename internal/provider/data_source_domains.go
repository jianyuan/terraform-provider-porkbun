package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
	"github.com/jianyuan/terraform-provider-porkbun/internal/fwdiag"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DomainsDataSourceModel struct {
	Domains supertypes.SetNestedObjectValueOf[DomainModel] `tfsdk:"domains"`
}

func (m *DomainsDataSourceModel) FromAPI(ctx context.Context, domains []apiclient.Domain) (diags diag.Diagnostics) {
	m.Domains = supertypes.NewSetNestedObjectValueOfValueSlice(ctx, lo.Map(domains, func(item apiclient.Domain, _ int) DomainModel {
		var mm DomainModel
		diags.Append(mm.FromAPI(ctx, item)...)
		return mm
	}))
	return
}

func NewDomainsDataSource() datasource.DataSource {
	return &DomainsDataSource{}
}

var _ datasource.DataSource = &DomainsDataSource{}
var _ datasource.DataSourceWithConfigure = &DomainsDataSource{}

type DomainsDataSource struct {
	baseDataSource
}

func (d *DomainsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domains"
}

func (d *DomainsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = domainsSchema().GetDataSource(ctx)
}

func (d *DomainsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data DomainsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	params := &apiclient.GetDomainsParams{
		Start: new(int64(0)),
	}

	var domains []apiclient.Domain

	for {
		httpResp, err := d.client.GetDomainsWithResponse(
			ctx,
			params,
		)
		if err != nil {
			resp.Diagnostics.Append(fwdiag.NewClientReadError(err))
			return
		} else if !apiclient.IsOK(httpResp) {
			resp.Diagnostics.Append(fwdiag.NewClientReadHTTPResponseError(httpResp))
			return
		}

		domains = append(domains, httpResp.JSON200.Domains...)

		if len(httpResp.JSON200.Domains) == 0 {
			break
		}

		params.Start = new(int64(len(domains) + 1))
	}

	resp.Diagnostics.Append(data.FromAPI(ctx, domains)...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
