package w3sql

import (
	"strings"
)

func EqualSQLStrings(a, b string) bool {
	a = NormalizeSQLString(strings.TrimSpace(a))
	b = NormalizeSQLString(strings.TrimSpace(b))
	return a == b
}
