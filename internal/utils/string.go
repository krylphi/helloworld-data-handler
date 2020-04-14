package utils

import "strings"

// Concat is used for fast string concatenation without fmt.Sprintf.
// because in go v1.13 there is no way that strings.Builder.WriteString returns errs.
// nolint: gosec
func Concat(strs ...string) string {
	b := strings.Builder{}
	for _, s := range strs {
		b.WriteString(s)
	}
	return b.String()
}
