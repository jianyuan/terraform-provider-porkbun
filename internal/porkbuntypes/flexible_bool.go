package porkbuntypes

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// FlexibleBool handles booleans represented as true, false, 1, 0, "1", "0", "true", "false", "yes", "no".
type FlexibleBool bool

func (fb *FlexibleBool) UnmarshalJSON(data []byte) error {
	raw := string(data)

	// Handle native JSON booleans: true / false
	if raw == "true" {
		*fb = true
		return nil
	}
	if raw == "false" {
		*fb = false
		return nil
	}

	// Handle numeric integers: 1 / 0
	if raw == "1" {
		*fb = true
		return nil
	}
	if raw == "0" {
		*fb = false
		return nil
	}

	// Handle quoted strings: "1", "0", "true", "false", "yes", "no"
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		str = strings.TrimSpace(strings.ToLower(str))
		switch str {
		case "1", "true", "yes":
			*fb = true
			return nil
		case "0", "false", "no":
			*fb = false
			return nil
		}
	}

	return fmt.Errorf("invalid boolean value: %s", raw)
}

func (fb FlexibleBool) MarshalJSON() ([]byte, error) {
	if fb {
		return []byte("1"), nil
	}
	return []byte("0"), nil
}

func FlexibleBoolPointerValue(v *FlexibleBool) types.Bool {
	if v == nil {
		return types.BoolNull()
	}
	return types.BoolValue(bool(*v))
}
