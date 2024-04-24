package w3sql

import (
	"errors"
	"fmt"
)

type DeleteQuery struct {
	CompiledQueryParams
	Is        map[string]map[string]string
	In        map[string]map[string][]string
	AllValues map[string]bool
}

func (cs *compilerSession) compileDeletePair(
	field string,
	value any,
	isInsert bool,
	allValues map[string]bool,
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
	allValues[fmt.Sprint(v)] = true
	return alias, true, nil
}

func (q *Query) CompileDelete(
	sqlSyntax string,
	fieldmap map[string]map[string]string,
	transform ...ValueTransform,
) (*DeleteQuery, error) {
	if q.Delete == nil {
		return nil, nil
	}
	result := &DeleteQuery{
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
		Is:        map[string]map[string]string{},
		In:        map[string]map[string][]string{},
		AllValues: map[string]bool{},
	}
	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		params:    map[string]any{},
	}

	for _, del := range q.Delete {
		for field, v := range del {
			fm, ok := fieldmap[field]
			if !ok {
				return nil, errors.New("w3sql: no such field " + field)
			}
			for table, f := range fm {
				if f == "" {
					f = field
				}
				var vv []any
				if vv, ok = v.([]any); !ok {
					aliases := make([]string, len(vv))
					for i, v := range vv {
						alias, ok, err := cs.compileDeletePair(table+"_"+f, v, false, result.AllValues, transform...)
						if err != nil {
							return nil, err
						}
						if !ok {
							continue
						}
						aliases[i] = alias
					}
					mp := result.In[table]
					if mp == nil {
						mp = map[string][]string{}
						result.In[table] = mp
					}
					mp[f] = aliases
				} else {
					alias, ok, err := cs.compileDeletePair(table+"_"+f, v, false, result.AllValues, transform...)
					if err != nil {
						return nil, err
					}
					if !ok {
						continue
					}
					mp := result.Is[table]
					if mp == nil {
						mp = map[string]string{}
						result.Is[table] = mp
					}
					mp[f] = alias
				}
			}
		}
	}

	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}
