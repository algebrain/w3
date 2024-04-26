package w3sql

func EqualSQLStrings(a, b string) bool {
	a = NormalizeSQLString(a, true)
	b = NormalizeSQLString(b, true)
	return a == b
}
