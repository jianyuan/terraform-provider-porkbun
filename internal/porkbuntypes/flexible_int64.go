package porkbuntypes

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FlexibleInt64 handles int64 values represented as JSON numbers or strings (e.g., 123, "123").
type FlexibleInt64 int64

func (fi *FlexibleInt64) UnmarshalJSON(data []byte) error {
	// 1. Try unmarshaling directly as a native JSON number (int64)
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*fi = FlexibleInt64(num)
		return nil
	}

	// 2. Try unmarshaling as a quoted string (e.g., "123")
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		str = strings.TrimSpace(str)
		parsed, err := strconv.ParseInt(str, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid int64 string value: %s", str)
		}
		*fi = FlexibleInt64(parsed)
		return nil
	}

	return fmt.Errorf("invalid int64 value: %s", string(data))
}

func (fi FlexibleInt64) MarshalJSON() ([]byte, error) {
	return json.Marshal(int64(fi))
}

func FlexibleInt64PointerValue(v *FlexibleInt64) types.Int64 {
	if v == nil {
		return types.Int64Null()
	}
	return types.Int64Value(int64(*v))
}
