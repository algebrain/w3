package w3sql

import (
	"strings"
)

func EqualSQLStrings(a, b string) bool {
	a = NormalizeSQLString(a, true)
	b = NormalizeSQLString(b, true)
	return a == b
}

func showSpaces(s string) string {
	s = strings.ReplaceAll(s, " ", "_")
	s = strings.ReplaceAll(s, "\n", `\`)
	s = strings.ReplaceAll(s, "\t", `|`)
	return strings.ReplaceAll(s, "\r", `/`)
}
