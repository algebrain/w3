package w3sql

import (
	"errors"
	"fmt"
	"strings"
)

type compilerSession struct {
	sqlSyntax  string
	params     map[string]any
	fieldmap   map[string]string
	varCounter int
}

type RawCondition interface {
	compile(*compilerSession) (string, error)
	LowerStringValues(map[string]bool)
}

type AtomaryCondition struct {
	Col  string
	Type string
	Val  any
	Op   string
}

type CompoundCondition struct {
	Op    string
	Query []RawCondition
}

type SortQuery struct {
	Col string
	Dir string
}

type QueryParam struct {
	Name string
	Val  any
}

type Query struct {
	Limit  *int
	Offset *int
	Search RawCondition
	Sort   []SortQuery
	Insert *struct {
		Cols   []string
		Values [][]any
	}
	Update *struct {
		Cols   []string
		Values [][]any
	}
	Delete []any
	Params map[string]any //дополнительные параметры запроса, вне логики SQL
}

type CompiledQueryParams struct {
	SQLParams map[string]any //пары ключ - значение для db.Exec
	Params    map[string]any //дополнительные параметры запроса, вне логики SQL
}

func (cs *compilerSession) getSearchField(fname string, ftype string) (string, bool) {
	var (
		field string
		ok    bool
	)

	if field, ok = cs.fieldmap[fname]; !ok {
		return "", false
	}

	if field == "" {
		field = fname
	}

	switch ftype {
	case "date":
		{
			switch cs.sqlSyntax {
			case "sqlite":
				field = "date(" + field + ", 'unixepoch')" // sqlite
			case "postgres":
				// OLD: field = field + "::int4::abstime::date" //postgres
				field = "to_timestamp(" + field + ")::date"
			}
		}
	case "datetime":
		{
			switch cs.sqlSyntax {
			case "sqlite":
				field = "date(" + field + ", 'unixepoch')" // sqlite
			case "postgres":
				// OLD: field = field + "::int4::abstime::date" //postgres
				field = "to_timestamp(" + field + ")"
			}
		}
	}

	return field, ok
}

func (q *AtomaryCondition) compile(cs *compilerSession) (string, error) {
	switch q.Op {
	case "равен", "is":
		return cs.compileOperatorIS(q, false)
	case "or", "или":
		return cs.compileOperatorOR(q)
	case "<=":
		return cs.compileOperatorLESS(q, true, true, false)
	case ">=":
		return cs.compileOperatorLESS(q, false, true, false)
	case "<":
		return cs.compileOperatorLESS(q, true, false, false)
	case ">":
		return cs.compileOperatorLESS(q, false, false, false)
	case ">= или 0", ">= or 0":
		return cs.compileOperatorLESS(q, false, true, true)
	case "не равен", "not is", "is not":
		return cs.compileOperatorIS(q, true)
	case "между", "between":
		return cs.compileOperatorBETWEEN(q)
	case "reverse in":
		return cs.compileOperatorReverseIN(q)
	case "в списке", "in":
		return cs.compileOperatorIN(q, false)
	case "не в списке", "not in":
		return cs.compileOperatorIN(q, true)
	case "начинается с", "begins", "starts with":
		return cs.compileOperatorBEGINS(q, false, false)
	case "содержит", "contains":
		return cs.compileOperatorBEGINS(q, true, false)
	case "заканчивается на", "ends", "ends with":
		return cs.compileOperatorBEGINS(q, false, true)
	default:
		return "", errors.New("w3sql: operator '" + q.Op + "' is not supported")
	}
}

func (q *CompoundCondition) compile(cs *compilerSession) (string, error) {
	parts := make([]string, len(q.Query))
	for i, qp := range q.Query {
		s, err := qp.compile(cs)
		if err != nil {
			return "", err
		}
		parts[i] = s
	}

	logics := strings.ToUpper(q.Op)
	not := false
	if q.Op == "NOT" {
		logics = "AND"
		not = true
	}

	result := strings.Join(parts, fmt.Sprintf(" %s ", logics))
	if not {
		result = "NOT (" + result + ")"
	}

	return "(" + result + ")", nil
}

func (q *SortQuery) compile(cs *compilerSession) (string, error) {
	q.Dir = strings.ToUpper(q.Dir)
	if q.Dir != "ASC" && q.Dir != "DESC" {
		return "", errors.New("w3sql: direction '" + q.Dir + "' is not supported")
	}
	var (
		field string
		ok    bool
	)
	if field, ok = cs.fieldmap[q.Col]; !ok {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}
	if field == "" {
		field = q.Col
	}

	return fmt.Sprintf("%v %v", field, q.Dir), nil
}

func And(a, b RawCondition) RawCondition {
	return &CompoundCondition{
		Op:    "AND",
		Query: []RawCondition{a, b},
	}
}

func Or(a, b RawCondition) RawCondition {
	return &CompoundCondition{
		Op:    "OR",
		Query: []RawCondition{a, b},
	}
}

func Not(a RawCondition) RawCondition {
	return &CompoundCondition{
		Op:    "NOT",
		Query: []RawCondition{a},
	}
}

func (c AtomaryCondition) LowerStringValues(cols map[string]bool) {
	if _, ok := cols[c.Col]; !ok {
		return
	}
	if v, ok := c.Val.(string); !ok {
		panic("[w3sql.AtomaryCondition.LowerStringValues]: not a string value: " + fmt.Sprint(c.Val))
	} else {
		c.Val = strings.ToLower(v)
	}
}

func (c CompoundCondition) LowerStringValues(cols map[string]bool) {
	for _, x := range c.Query {
		x.LowerStringValues(cols)
	}
}

func (q *Query) LowerSearchValues(cols map[string]bool) {
	q.Search.LowerStringValues(cols)
}
