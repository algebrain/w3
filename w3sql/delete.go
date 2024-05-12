package w3sql

import (
	"fmt"
)

type DeletePair struct {
	TableName string
	IDName    string
}

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
	var err error

	for _, fn := range transform {
		v, err = fn(tableName, idName, v)
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
	tables []*DeletePair,
	transform ...DeleteTransform,
) (*DeleteQuery, error) {
	if q.Delete == nil {
		return nil, nil
	}
	result := &DeleteQuery{
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
		Tables: make([]*TableDelete, len(tables)),
	}

	for i, tab := range tables {
		result.Tables[i] = &TableDelete{
			IDName:    tab.IDName,
			TableName: tab.TableName,
		}

		cs := &compilerSession{
			sqlSyntax: sqlSyntax,
			params:    map[string]any{},
		}
		result.Tables[i].ToDelete = make([]string, 0, len(q.Delete))

		for _, id := range q.Delete {
			alias, ok, err := cs.compileDeletePair(tab.TableName, tab.IDName, id, transform...)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			result.Tables[i].ToDelete = append(result.Tables[i].ToDelete, ":"+alias)
		}
		result.Tables[i].SQLParams = cs.params
	}

	return result, nil
}
