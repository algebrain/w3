package w3sql

import (
	"errors"
	"fmt"
)

type InsertPair struct {
	Fields []string
	Values []any
}

type InsertQuery struct {
	CompiledQueryParams
	Pairs []InsertPair
}

type UpdatePair struct {
	Field string
	Value any
}

type UpdateQuery struct {
	CompiledQueryParams
	IDField string
	Pairs   map[string][]UpdatePair
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

type ValueTransform func(isInsert bool, field string, value any) (any, error)

func (cs *compilerSession) compileWritePair(
	field string,
	value any,
	isInsert bool,
	transform ...ValueTransform,
) (string, bool, error) {
	v := value
	for _, fn := range transform {
		v, err := fn(isInsert, field, v)
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
		Pairs: make([]InsertPair, 0, len(q.Insert)),
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
	}
	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		fieldmap:  fieldmap,
		params:    map[string]any{},
	}

	idField, ok := cs.fieldmap["$id"]
	if !ok || idField == "" {
		idField = "id"
	}

	for _, ins := range q.Insert {
		pair := &InsertPair{
			Fields: make([]string, 0, len(ins)),
			Values: make([]any, 0, len(ins)),
		}
		for field, f := range fieldmap {
			if field == "$id" {
				continue
			}
			if f == "" {
				f = field
			}
			v, ok := ins[f]
			if !ok {
				v = nil
			}
			alias, ok, err := cs.compileWritePair(f, v, true, transform...)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			if f == idField && IsDefaultValue(cs.params[alias]) {
				continue
			}
			pair.Fields = append(pair.Fields, f)
			pair.Values = append(pair.Values, ":"+alias)
		}
		if len(pair.Fields) > 0 {
			result.Pairs = append(result.Pairs, *pair)
		}
	}

	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}

func (q *Query) CompileUpdate(
	sqlSyntax string,
	fieldmap map[string]string,
	transform ...ValueTransform,
) (*UpdateQuery, error) {
	if q.Update == nil {
		return nil, nil
	}
	result := &UpdateQuery{
		Pairs: map[string][]UpdatePair{},
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
	}
	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		fieldmap:  fieldmap,
		params:    map[string]any{},
	}

	idField, ok := cs.fieldmap["$id"]
	if !ok || idField == "" {
		idField = "id"
	}

	result.IDField = idField

	for _, upd := range q.Update {
		pairs := make([]UpdatePair, 0, len(upd)-1)
		id, ok := upd["$id"]
		if !ok {
			return nil, errors.New("w2ui: wrong update parameters")
		}
		for field, f := range fieldmap {
			if field == "$id" || idField == f {
				continue
			}
			if f == "" {
				f = field
			}
			v, ok := upd[f]
			if !ok {
				v = nil
			}
			alias, ok, err := cs.compileWritePair(f, v, false, transform...)
			if err != nil {
				return nil, err
			}
			if !ok {
				continue
			}
			pairs = append(pairs, UpdatePair{
				Field: f,
				Value: ":" + alias,
			})
		}
		if len(pairs) > 0 {
			alias := fmt.Sprintf("uiid%d", cs.varCounter)
			cs.varCounter++
			cs.params[alias] = id
			result.Pairs[alias] = pairs
		}
	}

	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}
