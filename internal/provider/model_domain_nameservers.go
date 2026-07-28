package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/samber/lo"
)

type DomainNameserversModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.Set    `tfsdk:"nameservers"`
}

func (m *DomainNameserversModel) Fill(ctx context.Context, nameservers []string) (diags diag.Diagnostics) {
	m.Nameservers = types.SetValueMust(types.StringType, lo.Map(nameservers, func(v string, _ int) attr.Value {
		return types.StringValue(v)
	}))
	return
}
