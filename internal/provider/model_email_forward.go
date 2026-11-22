package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

type EmailForwardModel struct {
	Address   types.String `tfsdk:"address"`
	ForwardTo types.String `tfsdk:"forward_to"`
	Created   types.String `tfsdk:"created"`
}

func (m *EmailForwardModel) FromAPI(ctx context.Context, data apiclient.GetEmailForwards200JSONResponseBody_Forwards) (diags diag.Diagnostics) {
	if data.Address == nil {
		m.Address = types.StringNull()
	} else {
		m.Address = types.StringValue(normalizeEmailForwardAddress(*data.Address))
	}
	m.ForwardTo = types.StringPointerValue(data.ForwardTo)
	m.Created = types.StringPointerValue(data.Created)
	return
}

type EmailForwardLimitsModel struct {
	Used types.Int64 `tfsdk:"used"`
	Max  types.Int64 `tfsdk:"max"`
}

func (m *EmailForwardLimitsModel) FromAPI(ctx context.Context, data apiclient.GetEmailForwards200JSONResponseBody_Limits) (diags diag.Diagnostics) {
	m.Used = types.Int64PointerValue(data.Used)
	m.Max = types.Int64PointerValue(data.Max)
	return
}

// normalizeEmailForwardAddress returns the local part (the part before the @)
// of an address. The API accepts both `info` and `info@example.com`.
func normalizeEmailForwardAddress(address string) string {
	address = strings.TrimSpace(address)
	if i := strings.LastIndex(address, "@"); i >= 0 {
		return address[:i]
	}
	return address
}

// emailForwardMatches reports whether an API forward matches the given
// address and destination. Email addresses are compared case-insensitively.
func emailForwardMatches(apiAddress, apiForwardTo, address, forwardTo string) bool {
	return strings.EqualFold(normalizeEmailForwardAddress(apiAddress), normalizeEmailForwardAddress(address)) &&
		strings.EqualFold(strings.TrimSpace(apiForwardTo), strings.TrimSpace(forwardTo))
}
