package w3sql

import (
	"fmt"
	"regexp"
	"strings"
)

type SQLQuery struct {
	Code   string
	Params map[string]any

	Base       string
	Conditions string
	Limit      string
	Offset     string
	Order      string
	Fields     string
	Values     string
}

func removeRoundBracketsContents(s string) string {
	str := []rune(s)
	values := make([]rune, 0, len(str))
	bracket := 0
	isStringLiteral := false
	isEscape := false

	for i := 0; i < len(str); i++ {
		ch := str[i]
		needEscape := false

		switch ch {
		case '(':
			if !isStringLiteral {
				bracket++
				continue
			}
		case ')':
			if !isStringLiteral {
				bracket--
				continue
			}
		case '\'':
			if !isStringLiteral {
				isStringLiteral = true
			} else if !isEscape {
				isStringLiteral = false
			}
		case '\\':
			if isStringLiteral {
				needEscape = true
			}
		}

		isEscape = needEscape

		if bracket == 0 {
			values = append(values, ch)
		}
	}

	return string(values)
}

var spaces = regexp.MustCompile(`\s+`)

func NormalizeSQLString(s string, trimmed ...bool) string {
	s = strings.ToLower(s)
	if len(trimmed) > 0 && trimmed[0] {
		s = strings.TrimSpace(s)
	}
	return spaces.ReplaceAllLiteralString(s, " ")
}

func NeedsWhere(baseSQL string) bool {
	s := NormalizeSQLString(baseSQL)
	s = removeRoundBracketsContents(s)

	where := strings.LastIndex(s, " where ")
	from := strings.LastIndex(s, " from ")
	join := strings.LastIndex(s, " join ")

	return where < from || where < join || where == -1
}

func (cq *SelectQuery) SQL(baseSQL ...*SQLString) ([]SQLQuery, error) {
	result := SQLQuery{Params: cq.SQLParams}
	needsWhere := true

	if baseSQL != nil && len(baseSQL) > 0 {
		result.Base = baseSQL[0].String()
		result.Code += result.Base
		needsWhere = baseSQL[0].NeedsWhere()
	}
	if cq.Conditions != "" {
		result.Conditions = cq.Conditions
		if needsWhere { // where ... from ... join
			result.Code += "\nwhere " + cq.Conditions + " " // оставляем  [where ... from ... join] + [where ...]
		} else {
			result.Code += "\nand " + cq.Conditions + " "
		}
	}

	if cq.Limit != nil {
		result.Limit = fmt.Sprintf("limit %d ", *cq.Limit)
		result.Code += "\n" + result.Limit
	}

	if cq.Offset != nil {
		result.Offset = fmt.Sprintf("offset %d ", *cq.Offset)
		result.Code += "\n" + result.Offset
	}

	if cq.Order != nil && len(cq.Order) > 0 {
		result.Order = "order by " + strings.Join(cq.Order, ", ")
		result.Code += "\n" + result.Order
	}

	return []SQLQuery{result}, nil
}

func (cq *SelectQuery) NoLimitOffset() *SelectQuery {
	result := *cq
	result.Offset = nil
	result.Limit = nil
	return &result
}

func (cq *SelectQuery) NoOrder() *SelectQuery {
	result := *cq
	result.Order = nil
	return &result
}

func (cq *SelectQuery) NoConditions() *SelectQuery {
	result := *cq
	result.Conditions = ""
	return &result
}

func (q *InsertQuery) SQL(baseSQL ...*SQLString) ([]SQLQuery, error) {
	result := SQLQuery{Params: q.SQLParams}
	if baseSQL != nil && len(baseSQL) > 0 {
		result.Base = baseSQL[0].String()
		result.Code += result.Base
	}
	result.Fields = fmt.Sprintf(" (%s)", strings.Join(q.Fields, ","))
	result.Code += result.Fields
	vals := make([]string, len(q.Values))
	for i, v := range q.Values {
		vals[i] = "(" + strings.Join(v, ",") + ")"
	}
	result.Values = "values\n" + strings.Join(vals, ",\n")
	result.Code += "\n" + result.Values
	return []SQLQuery{result}, nil
}

func (q *UpdateQuery) SQL(baseSQL ...*SQLString) ([]SQLQuery, error) {
	result := SQLQuery{Params: q.SQLParams}
	if baseSQL != nil && len(baseSQL) > 0 {
		result.Base = baseSQL[0].String()
		result.Code += result.Base
	}

	flds := make([]string, len(q.Fields))
	for i, f := range q.Fields {
		if f == q.IDField {
			continue
		}
		flds[i] = fmt.Sprintf("%s = c.%s", f, f)
	}
	result.Fields = " set\n" + strings.Join(flds, ",\n")
	result.Code += result.Fields

	vals := make([]string, len(q.Values))
	for i, v := range q.Values {
		vals[i] = "(" + strings.Join(v, ",") + ")"
	}
	result.Values = "from (values \n" + strings.Join(vals, ",\n") + "\n) as c"
	result.Values += "(" + strings.Join(q.Fields, ",") + ")"
	result.Code += result.Values

	result.Conditions = fmt.Sprintf("where %s = c.%s", q.IDField, q.IDField)
	result.Code += "\n" + result.Conditions

	return []SQLQuery{result}, nil
}

type IsDelAllowedFunc = func(any) error

func (q *DeleteQuery) SQL(baseSQL ...*SQLString) ([]SQLQuery, error) {
	result := make([]SQLQuery, len(q.Tables))
	if baseSQL != nil {
		for i, bs := range baseSQL {
			if i >= len(result) {
				break
			}
			result[i].Code = bs.String() + "\n"
		}
	}

	for i, tab := range q.Tables {
		if len(tab.ToDelete) == 1 {
			result[i].Code = fmt.Sprintf(
				"delete from %s where %s = %s",
				tab.TableName,
				tab.IDName,
				tab.ToDelete[0],
			)
		} else {
			result[i].Code = fmt.Sprintf(
				"delete from %s where %s in (%s)",
				tab.TableName,
				tab.IDName,
				strings.Join(tab.ToDelete, ","),
			)
		}
		result[i].Params = tab.SQLParams
	}

	return result, nil
}
