package w3sql

import (
	"strings"
)

func equalSQLStrings(a, b string) bool {
	a = normalizeSQLString(strings.TrimSpace(a))
	b = normalizeSQLString(strings.TrimSpace(b))
	return a == b
}
