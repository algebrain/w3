package w3ui

type SQLSyntax string

const (
	SyntaxPostgreSQL SQLSyntax = "postgres"
	SyntaxSQLite     SQLSyntax = "sqlite"
)

type LogPurpose string

type GlobalConfig struct {
	SQLSyntax SQLSyntax
	GetLogger func(requestPath string, logPurpose LogPurpose) ExtLogger
}

var globalConfig = GlobalConfig{
	SQLSyntax: SyntaxSQLite,
	GetLogger: nil,
}

func SetGlobalConfig(cfg GlobalConfig) {
	globalConfig = cfg
}

// "postgres" / "sqlite"
func SetSQLSyntax(s SQLSyntax) {
	globalConfig.SQLSyntax = s
}
