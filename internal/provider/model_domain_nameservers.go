package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	supertypes "github.com/orange-cloudavenue/terraform-plugin-framework-supertypes"
	"github.com/samber/lo"
)

type DomainNameserversModel struct {
	Domain      types.String                  `tfsdk:"domain"`
	Nameservers supertypes.SetValueOf[string] `tfsdk:"nameservers"`
}

func (m *DomainNameserversModel) Fill(ctx context.Context, nameservers []string) (diags diag.Diagnostics) {
	diags.Append(m.Nameservers.Set(ctx, lo.Uniq(nameservers))...)
	return
}
