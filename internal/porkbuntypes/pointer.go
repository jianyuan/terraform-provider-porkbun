package porkbuntypes

// StringPointerAs converts a *string into a pointer to a string-based type T
// (for example a generated enum). A nil input yields nil.
func StringPointerAs[T ~string](v *string) *T {
	if v == nil {
		return nil
	}
	t := T(*v)
	return &t
}
