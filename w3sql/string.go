package w3sql

type SQLString struct {
	data       string
	needsWhere bool
}

// тип SQLString создан для большей производительности запросов
// метод NewSQLString следует вызывать заранее, вне тела запроса
func NewSQLString(s string) SQLString {
	return SQLString{
		data:       s,
		needsWhere: NeedsWhere(s),
	}
}

func (s SQLString) String() string {
	return s.data
}

func (s SQLString) NeedsWhere() bool {
	return s.needsWhere
}
