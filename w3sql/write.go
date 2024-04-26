package w3sql

import (
	"errors"
	"fmt"
)

type InsertQuery struct {
	CompiledQueryParams
	Fields []string
	Values [][]string
}

type UpdateQuery struct {
	CompiledQueryParams
	IDField string
	Fields  []string
	Values  [][]string
}

func IsDefaultValue(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case int, uint, int32, uint32, int64, uint64:
		return x == 0
	case float32, float64:
		return x == 0
	case string:
		return x == ""
	case []byte:
		return len(x) == 0
	}
	return false
}

type ValueTransform func(field string, value any) (any, error)

func (cs *compilerSession) compileWritePair(
	field string,
	value any,
	transform ...ValueTransform,
) (string, bool, error) {
	v := value
	var err error
	for _, fn := range transform {
		v, err = fn(field, v)
		if err != nil {
			return "", false, err
		}
		if v == nil {
			return "", false, nil
		}
	}

	alias := fmt.Sprintf("ui%s%d", field, cs.varCounter)
	cs.varCounter++
	cs.params[alias] = v
	return alias, true, nil
}

func (q *Query) CompileInsert(
	sqlSyntax string,
	fieldmap map[string]string,
	transform ...ValueTransform,
) (*InsertQuery, error) {
	if q.Insert == nil {
		return nil, nil
	}
	result := &InsertQuery{
		Fields: make([]string, len(q.Insert.Fields)),
		Values: make([][]string, 0, len(q.Insert.Values)),
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
	}
	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		fieldmap:  fieldmap,
		params:    map[string]any{},
	}

	for i, field := range q.Insert.Fields {
		f, ok := fieldmap[field]
		if !ok {
			return nil, errors.New("w3sql: no such field: " + field)
		}
		if f == "" {
			f = field
		}
		result.Fields[i] = f
	}

rows:
	for i, vals := range q.Insert.Values {
		if len(vals) != len(result.Fields) {
			return nil, errors.New("w3sql: wrong length of list of values in position " + fmt.Sprint(i))
		}
		rVals := make([]string, len(vals))
		for j, v := range vals {
			field := q.Insert.Fields[j]
			alias, ok, err := cs.compileWritePair(field, v, transform...)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue rows
			}
			rVals[j] = ":" + alias
		}
		result.Values = append(result.Values, rVals)
	}

	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}

func (q *Query) CompileUpdate(
	sqlSyntax string,
	fieldmap map[string]string,
	idFieldName string,
	transform ...ValueTransform,
) (*UpdateQuery, error) {
	if q.Update == nil {
		return nil, nil
	}
	result := &UpdateQuery{
		Fields:  make([]string, len(q.Update.Fields)),
		Values:  make([][]string, 0, len(q.Update.Values)),
		IDField: idFieldName,
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
	}
	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		fieldmap:  fieldmap,
		params:    map[string]any{},
	}

	idFound := false
	for i, field := range q.Update.Fields {
		f, ok := fieldmap[field]
		if !ok {
			return nil, errors.New("w3sql: no such field: " + field)
		}
		if f == "" {
			f = field
		}
		if f == idFieldName {
			idFound = true
		}
		result.Fields[i] = f
	}

	if !idFound {
		return nil, errors.New("w3sql: id not found")
	}

rows:
	for i, vals := range q.Update.Values {
		if len(vals) != len(result.Fields) {
			return nil, errors.New("w3sql: wrong length of list of values in position " + fmt.Sprint(i))
		}
		rVals := make([]string, len(vals))
		for j, v := range vals {
			field := q.Update.Fields[j]
			alias, ok, err := cs.compileWritePair(field, v, transform...)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue rows
			}
			rVals[j] = ":" + alias
		}
		result.Values = append(result.Values, rVals)
	}

	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}
