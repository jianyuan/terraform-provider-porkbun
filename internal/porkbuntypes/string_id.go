package porkbuntypes

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// StringID represents an ID that can be unmarshaled from either a JSON string or a JSON number,
// but always marshals back into a JSON string.
type StringID string

func (id *StringID) UnmarshalJSON(data []byte) error {
	// First, attempt to decode as a JSON string
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*id = StringID(s)
		return nil
	}

	// Next, attempt to decode as a JSON number (int64)
	var num int64
	if err := json.Unmarshal(data, &num); err == nil {
		*id = StringID(strconv.FormatInt(num, 10))
		return nil
	}

	return fmt.Errorf("StringID: cannot unmarshal %s into string or integer", string(data))
}

func (id StringID) MarshalJSON() ([]byte, error) {
	// Always output as a quoted JSON string
	return json.Marshal(string(id))
}

func StringIDPointerValue(v *StringID) types.String {
	if v == nil {
		return types.StringNull()
	}
	return types.StringValue(string(*v))
}
