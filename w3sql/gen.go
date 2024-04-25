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
	result := make([]string, len(q.Pairs))
	for i, pair := range q.Pairs {
		values := make([]string, len(pair.Values))
		for j, v := range pair.Values {
			values[j] = fmt.Sprint(v)
		}
		result[i] = baseSQL.String() +
			fmt.Sprintf(" (%s)", strings.Join(pair.Fields, ",")) +
			fmt.Sprintf(" values (%s)", strings.Join(values, ","))
	}
	return result, q.SQLParams, nil
}

func (q *UpdateQuery) SQL(baseSQL *SQLString) ([]string, map[string]any, error) {
	result := make([]string, 0, len(q.Pairs))
	for k, pairs := range q.Pairs {
		updates := make([]string, len(pairs))
		for j, p := range pairs {
			updates[j] = fmt.Sprintf("%s=%v", p.Field, p.Val)
		}
		result = append(result, baseSQL.String()+" "+
			strings.Join(updates, ",")+
			fmt.Sprintf(" where %s=%s;\n", q.IDField, ":"+k),
		)
	}
	return result, q.SQLParams, nil
}

type IsDelAllowedFunc = func(any) error

func (q *DeleteQuery) SQL(baseSQL *SQLString, isDelAllowed ...IsDelAllowedFunc) ([]string, map[string]any, error) {
	if len(isDelAllowed) > 0 {
		for v, _ := range q.AllValues {
			if err := isDelAllowed[0](v); err != nil {
				return nil, nil, err
			}
		}
	}
	result := make([]string, 0, 10)
	isBaseAdded := false
	prefix := baseSQL.String()
	if prefix != "" {
		prefix += "\n"
	}
	for table, deletes := range q.Is {
		for field, val := range deletes {
			ns := fmt.Sprintf("delete from %s where %s=%s", table, field, val)
			if !isBaseAdded {
				ns = prefix + ns
				isBaseAdded = true
			}
			result = append(result, ns)
		}
	}

	for table, deletes := range q.In {
		for field, val := range deletes {
			ns := fmt.Sprintf(
				"delete from %s where %s in (%s)",
				table,
				field,
				strings.Join(val, ","),
			)
			if !isBaseAdded {
				ns = prefix + ns
				isBaseAdded = true
			}
			result = append(result, ns)
		}
	}

	return result, q.SQLParams, nil
}
