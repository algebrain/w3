package w3sql

import (
	"fmt"
)

type TablePair struct {
	TableName string
	IDName    string
}

type DeleteQuery struct {
	CompiledQueryParams
	ToDelete  []string
	Tables    []TablePair
	AllValues map[string]bool
}

type DeleteTransform func(id any) (any, error)

func (cs *compilerSession) compileDeletePair(
	id any,
	allValues map[string]bool,
	transform ...DeleteTransform,
) (string, bool, error) {
	v := id
	for _, fn := range transform {
		v, err := fn(v)
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
	allValues[fmt.Sprint(v)] = true
	return alias, true, nil
}

func (q *Query) CompileDelete(
	sqlSyntax string,
	tables []TablePair,
	transform ...DeleteTransform,
) (*DeleteQuery, error) {
	if q.Delete == nil {
		return nil, nil
	}
	result := &DeleteQuery{
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
		ToDelete:  make([]string, len(q.Delete)),
		Tables:    tables,
		AllValues: map[string]bool{},
	}

	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		params:    map[string]any{},
	}

	for i, id := range q.Delete {
		alias, ok, err := cs.compileDeletePair(id, result.AllValues, transform...)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		result.ToDelete[i] = ":" + alias
	}

	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}
