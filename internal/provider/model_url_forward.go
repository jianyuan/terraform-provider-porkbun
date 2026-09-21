package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type URLForwardModel struct {
	Id           types.String `tfsdk:"id"`
	Subdomain    types.String `tfsdk:"subdomain"`
	IncludePath  types.Bool   `tfsdk:"include_path"`
	Location     types.String `tfsdk:"location"`
	Type         types.String `tfsdk:"type"`
	RedirectType types.String `tfsdk:"redirect_type"`
	Wildcard     types.Bool   `tfsdk:"wildcard"`
}

func (m *URLForwardModel) FromAPI(ctx context.Context, data apiclient.GetUrlForwardingResponse_Forwards) (diags diag.Diagnostics) {
	m.Id = types.StringPointerValue(data.Id)

	if data.IncludePath == nil {
		m.IncludePath = types.BoolNull()
	} else {
		m.IncludePath = types.BoolValue(*data.IncludePath == apiclient.GetUrlForwardingResponseForwardsIncludePathYes)
	}

	m.Location = types.StringPointerValue(data.Location)

	if data.RedirectType == nil {
		m.RedirectType = types.StringNull()
	} else {
		m.RedirectType = types.StringValue(string(*data.RedirectType))
	}

	if data.Subdomain == nil || *data.Subdomain == "" {
		m.Subdomain = types.StringNull()
	} else {
		m.Subdomain = types.StringValue(*data.Subdomain)
	}

	if data.Type == nil {
		m.Type = types.StringNull()
	} else {
		m.Type = types.StringValue(string(*data.Type))
	}

	if data.Wildcard == nil {
		m.Wildcard = types.BoolNull()
	} else {
		m.Wildcard = types.BoolValue(*data.Wildcard == apiclient.GetUrlForwardingResponseForwardsWildcardYes)
	}

	return
}
