package porkbuntypes

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func StringIntegerBoolValue(v string) (types.Bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch v {
	case "0":
		return types.BoolValue(false), diags
	case "1":
		return types.BoolValue(true), diags
	default:
		diags.AddError(fmt.Sprintf("invalid value %q", v), fmt.Sprintf("invalid value %q", v))
		return types.BoolUnknown(), diags
	}
}

func IntegerBoolValue[T int | int32 | int64](v T) (types.Bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	switch v {
	case 0:
		return types.BoolValue(false), diags
	case 1:
		return types.BoolValue(true), diags
	default:
		diags.AddError(fmt.Sprintf("invalid value %q", v), fmt.Sprintf("invalid value %q", v))
		return types.BoolUnknown(), diags
	}
}
