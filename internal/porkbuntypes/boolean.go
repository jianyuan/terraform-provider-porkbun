package porkbuntypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/jianyuan/terraform-provider-porkbun/internal/apiclient"
)

func BoolValue(v interface {
	AsBoolInteger() (apiclient.BoolInteger, error)
	AsBoolString() (apiclient.BoolString, error)
}) (types.Bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	if vInt, err := v.AsBoolInteger(); err == nil {
		return types.BoolValue(vInt == 1), diags
	} else if vStr, err := v.AsBoolString(); err == nil {
		return types.BoolValue(vStr == "1"), diags
	} else {
		diags.AddError(fmt.Sprintf("invalid value %q", v), fmt.Sprintf("invalid value %q", v))
		return types.BoolUnknown(), diags
	}
}
