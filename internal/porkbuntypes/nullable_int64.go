package porkbuntypes

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

func NullableInt64Value(value nullable.Nullable[int64]) types.Int64 {
	if v, err := value.Get(); err == nil {
		return types.Int64Value(v)
	}
	return types.Int64Null()

}
