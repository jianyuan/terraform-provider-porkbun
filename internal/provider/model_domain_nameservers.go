package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DomainNameserversModel struct {
	Domain      types.String                  `tfsdk:"domain"`
	Nameservers supertypes.SetValueOf[string] `tfsdk:"nameservers"`
}

func (m *DomainNameserversModel) FromAPI(ctx context.Context, nameservers []string) (diags diag.Diagnostics) {
	normalized := lo.Map(nameservers, func(ns string, _ int) string {
		return strings.TrimSuffix(ns, ".")
	})
	diags.Append(m.Nameservers.Set(ctx, lo.Uniq(normalized))...)
	return
}
