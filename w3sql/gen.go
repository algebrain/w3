package w3sql

import (
	"fmt"
	"regexp"
	"strings"
)

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

func normalizeSQLString(s string) string {
	s = strings.ToLower(s)
	return spaces.ReplaceAllLiteralString(s, " ")
}

func NeedsWhere(baseSQL string) bool {
	s := normalizeSQLString(baseSQL)
	s = removeRoundBracketsContents(s)

	where := strings.LastIndex(s, " where ")
	from := strings.LastIndex(s, " from ")
	join := strings.LastIndex(s, " join ")

	return where < from || where < join || where == -1
}

func (cq *SelectQuery) SQL(baseSQL *SQLString) ([]string, map[string]any, error) {
	result := baseSQL.String()

	if cq.Conditions != "" {
		if baseSQL.NeedsWhere() { // where ... from ... join
			result += "\nwhere " + cq.Conditions + " " // оставляем  [where ... from ... join] + [where ...]
		} else {
			result += "\nand " + cq.Conditions + " "
		}
	}

	if cq.Limit != nil {
		result += fmt.Sprintf("\nlimit %d ", *cq.Limit)
	}

	if cq.Offset != nil {
		result += fmt.Sprintf("\noffset %d ", *cq.Offset)
	}

	if cq.Order != nil && len(cq.Order) > 0 {
		result += "\norder by " + strings.Join(cq.Order, ", ")
	}

	return []string{result}, cq.SQLParams, nil
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

func (q *InsertQuery) SQL(baseSQL *SQLString) ([]string, map[string]any, error) {
	result := baseSQL.String() +
		fmt.Sprintf(" (%s)", strings.Join(q.Fields, ",")) +
		"\nvalues\n"
	vals := make([]string, len(q.Values))
	for i, v := range q.Values {
		vals[i] = "(" + strings.Join(v, ",") + ")"
	}
	result += strings.Join(vals, ",\n")
	return []string{result}, q.SQLParams, nil
}

func (q *UpdateQuery) SQL(baseSQL *SQLString) ([]string, map[string]any, error) {
	result := baseSQL.String() + " set\n"
	flds := make([]string, len(q.Fields))
	for i, f := range q.Fields {
		if f == q.IDField {
			continue
		}
		flds[i] = fmt.Sprintf("%s = c.%s", f, f)
	}
	result += strings.Join(flds, ",\n")
	vals := make([]string, len(q.Values))
	for i, v := range q.Values {
		vals[i] = "(" + strings.Join(v, ",") + ")"
	}
	result += "from (values \n" + strings.Join(vals, ",\n") + "\n) as c"
	result += "(" + strings.Join(q.Fields, ",") + ")"
	result += fmt.Sprintf("\n where %s = c.%s", q.IDField, q.IDField)
	return []string{result}, q.SQLParams, nil
}

type IsDelAllowedFunc = func(any) error

func (q *DeleteQuery) SQL(baseSQL *SQLString) ([]string, map[string]any, error) {
	result := make([]string, 0, len(q.Tables))
	prefix := baseSQL.String()
	isBaseAdded := prefix == ""

	for _, tab := range q.Tables {
		ns := fmt.Sprintf(
			"delete from %s where %s in (%s)",
			tab.TableName,
			tab.IDName,
			strings.Join(q.ToDelete, ","),
		)
		if !isBaseAdded {
			ns = prefix + "\n" + ns
			isBaseAdded = true
		}
		result = append(result, ns)
	}

	return result, q.SQLParams, nil
}
