package porkbuntypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

func NullableStringValue(value nullable.Nullable[string]) types.String {
	if v, err := value.Get(); err == nil {
		return types.StringValue(v)
	}
	return types.StringNull()

}
