package w3ui

import (
	"errors"

	"github.com/algebrain/w3/w3sql"
)

type Query w3sql.Query

type SqlQuery struct {
	Text string

	offset  int64
	limit   int64  // по умолчанию, либо переданный через Compile
	NoLimit string //запрос что и Text , но без order offset limit

	Sort string
	Map  map[string]any // пары ключ - значение для db.Exec
}

func (q *Query) Compile(fieldmap map[string]string) (*SqlQuery, error) {
	sq, err := (*w3sql.Query)(q).CompileSelect(string(globalConfig.SQLSyntax), fieldmap)
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

type SqlUpsertQuery struct {
	InsertQueries []string
	InsertValues  map[string]any
	UpdateQueries map[string]string
	UpdateValues  map[string]any
}

func (q *Query) CompileUpsert(
	idFieldName string,
	fieldmap map[string]bool,
	fn func(bool, string, any) (any, error),
) (*SqlUpsertQuery, error) {

	result := &SqlUpsertQuery{
		UpdateQueries: map[string]string{},
	}

	fu := func(field string, value any) (any, error) {
		return fn(false, field, value)
	}

	fm := map[string]string{}
	for k, _ := range fieldmap {
		fm[k] = ""
	}

	uq, err := (*w3sql.Query)(q).CompileUpdate(string(globalConfig.SQLSyntax), fm, idFieldName, fu)
	if err != nil {
		return nil, err
	}

	if uq != nil {
		usql, err := uq.SQL()
		if err != nil {
			return nil, err
		}

		result.UpdateQueries[idFieldName] = usql[0].Code
		result.UpdateValues = usql[0].Params
	}

	fi := func(field string, value any) (any, error) {
		return fn(true, field, value)
	}

	iq, err := (*w3sql.Query)(q).CompileInsert(string(globalConfig.SQLSyntax), fm, fi)
	if err != nil {
		return nil, err
	}

	if iq != nil {
		isql, err := iq.SQL()
		if err != nil {
			return nil, err
		}

		result.InsertQueries = []string{isql[0].Code}
		result.InsertValues = isql[0].Params
	}

	return result, nil
}

type isDeletionAllowedFunc func(id string) (bool, string)

func (q *Query) CompileDeleteSql(
	tableIds map[string]string,
	isDeletionAllowed isDeletionAllowedFunc,
) ([]*SqlQuery, string) {

	tables := make([]*w3sql.DeletePair, 0, len(tableIds))
	for t, id := range tableIds {
		tables = append(tables, &w3sql.DeletePair{
			TableName: t,
			IDName:    id,
		})
	}

	fd := func(tableName string, idName string, id any) (any, error) {
		ok, errString := isDeletionAllowed(idName)
		if !ok {
			return nil, errors.New(errString)
		}
		return id, nil
	}

	dq, err := (*w3sql.Query)(q).CompileDelete(string(globalConfig.SQLSyntax), tables, fd)

	if err != nil {
		return nil, err.Error()
	}

	if dq == nil {
		return nil, ""
	}

	dsql, err := dq.SQL()
	if err != nil {
		return nil, err.Error()
	}

	result := make([]*SqlQuery, len(dsql))
	for i, s := range dsql {
		r := &SqlQuery{
			Text: s.Code,
			Map:  s.Params,
		}
		result[i] = r
	}

	return result, ""
}
