package w3ui

type SQLSyntax string

const (
	SyntaxPostgreSQL SQLSyntax = "postgres"
	SyntaxSQLite     SQLSyntax = "sqlite"
)

type LogPurpose string

type GlobalConfig struct {
	SQLSyntax  SQLSyntax
	GetLogger  func(requestPath string, logPurpose LogPurpose) ExtLogger
	ErrorCodes ErrorCodes
}

var globalConfig = GlobalConfig{
	SQLSyntax: SyntaxSQLite,
	GetLogger: nil,
	ErrorCodes: ErrorCodes{
		SYSTEM_ERROR:       1,
		INVALID_PARAMETERS: 2,
	},
}

func SetGlobalConfig(cfg GlobalConfig) {
	globalConfig = cfg
}

// "postgres" / "sqlite"
func SetSQLSyntax(s SQLSyntax) {
	globalConfig.SQLSyntax = s
}

func SetErrorCodes(m map[string]int) {
	for k, v := range m {
		globalConfig.ErrorCodes[k] = v
	}
}
