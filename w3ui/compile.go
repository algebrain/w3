package w3ui

import (
	"github.com/algebrain/w3/w3sql"
)

type Query w3sql.Query

type SqlQuery struct {
	Text string

	offset  int64
	limit   int64  // по умолчанию, либо переданный через Compile
	NoLimit string //запрос что и Text , но без order offset limit

	Sort string
	Map  map[string]interface{} // пары ключ - значение для db.Exec
}

func (q *Query) Compile(fieldmap map[string]string) (*SqlQuery, error) {
	sq, err := (*w3sql.Query)(q).CompileSelect(sqlSyntax, fieldmap)
	if err != nil {
		return nil, err
	}

	result := &SqlQuery{}
	qs, err := sq.SQL(nil)
	if err != nil {
		return nil, err
	}
	qq := qs[0]

	conds := ""
	if qq.Conditions != "" {
		conds = "where " + qq.Conditions
	}
	result.Text = joinNonEmpty([]string{qq.Base, conds, qq.Order, qq.Limit, qq.Offset}, "\n")
	if q.Offset != nil {
		result.offset = int64(*q.Offset)
	}
	if q.Limit != nil {
		result.limit = int64(*q.Limit)
	}
	result.Map = qs[0].Params

	result.NoLimit = joinNonEmpty([]string{qq.Base, conds}, "\n")
	result.Sort = qq.Order

	return result, nil
}
