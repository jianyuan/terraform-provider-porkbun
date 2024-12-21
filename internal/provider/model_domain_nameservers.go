package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/go-utils/sliceutils"
)

type DomainNameserversModel struct {
	Domain      types.String `tfsdk:"domain"`
	Nameservers types.Set    `tfsdk:"nameservers"`
}

func (m *DomainNameserversModel) Fill(ctx context.Context, nameservers []string) (diags diag.Diagnostics) {
	m.Nameservers = types.SetValueMust(types.StringType, sliceutils.Map(func(v string) attr.Value {
		return types.StringValue(v)
	}, nameservers))
	return
}
