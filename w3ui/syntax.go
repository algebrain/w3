package w3ui

var sqlSyntax string = "postgres"

// "postgres" / "sqlite"
func SelectSQLSyntax(name string) {
	sqlSyntax = name
}
