package w3sql

import (
	"fmt"
)

type TableDelete struct {
	TableName string
	IDName    string
	ToDelete  []string
	SQLParams map[string]any
}

type DeleteQuery struct {
	CompiledQueryParams
	Tables []*TableDelete
}

type DeleteTransform func(tableName string, idName string, id any) (any, error)

func (cs *compilerSession) compileDeletePair(
	tableName string,
	idName string,
	id any,
	transform ...DeleteTransform,
) (string, bool, error) {
	v := id

	for _, fn := range transform {
		v, err := fn(tableName, idName, v)
		if err != nil {
			return "", false, err
		}
		if v == nil {
			return "", false, nil
		}
	}

	alias := fmt.Sprintf("ui%d", cs.varCounter)
	cs.varCounter++
	cs.params[alias] = v
	return alias, true, nil
}

func (q *Query) CompileDelete(
	sqlSyntax string,
	tables []*TableDelete,
	transform ...DeleteTransform,
) (*DeleteQuery, error) {
	if q.Delete == nil {
		return nil, nil
	}
	result := &DeleteQuery{
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
		Tables: tables,
	}

	for _, tab := range tables {
		cs := &compilerSession{
			sqlSyntax: sqlSyntax,
			params:    map[string]any{},
		}
		tab.ToDelete = make([]string, 0, len(q.Delete))

		for _, id := range q.Delete {
			alias, ok, err := cs.compileDeletePair(tab.TableName, tab.IDName, id, transform...)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			tab.ToDelete = append(tab.ToDelete, ":"+alias)
		}
		tab.SQLParams = cs.params
	}

	return result, nil
}
