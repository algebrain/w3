package w3sql

import (
	"errors"
	"fmt"
	"strings"
)

func (cs *compilerSession) compileOperatorIS(q *AtomaryCondition, not bool) (string, error) {
	vn := "sqv" + fmt.Sprint(cs.varCounter)
	cs.varCounter++
	result := ""
	var err error
	if field, ok := cs.getSearchField(q.Col, q.Type); ok {
		op := "="
		if not {
			op = "<>"
		}
		result = fmt.Sprintf("(%v%s:%v)", field, op, vn)
	} else {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}
	cs.params[vn], err = convValue(q.Val, q.Type)
	return result, err
}

func (cs *compilerSession) compileOperatorOR(q *AtomaryCondition) (string, error) {
	var err error
	sq, ok := q.Val.([]any)
	if !ok {
		return "", errors.New("w3sql: unexpected value of parameter" + q.Col)
	}
	parts := make([]string, len(sq))
	for j, val := range sq {
		vn := "sqv" + fmt.Sprintf("%d_a%d", cs.varCounter, j)
		if field, ok := cs.getSearchField(q.Col, q.Type); ok {
			parts[j] = fmt.Sprintf("(%v=:%v)", field, vn)
		} else {
			return "", errors.New("w3sql: no such field name " + q.Col)
		}
		cs.params[vn], err = convValueElem(val, q.Type)
		if err != nil {
			return "", err
		}
	}
	cs.varCounter++
	if len(parts) < 1 {
		return "", nil
	}
	return " (" + strings.Join(parts, " or ") + ") ", nil
}

func (cs *compilerSession) compileOperatorLESS(q *AtomaryCondition, less, orEqual, orZero bool) (string, error) {
	vn := "sqv" + fmt.Sprint(cs.varCounter)
	cs.varCounter++
	result := ""
	var err error
	if field, ok := cs.getSearchField(q.Col, q.Type); ok {
		op := ">"
		if less {
			op = "<"
		}
		if orEqual {
			op += "="
		}
		result = fmt.Sprintf("%v%s:%v", field, op, vn)
		if orZero {
			result += fmt.Sprintf("  or %v=0", field)
		}
		result = "(" + result + ")"
	} else {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}
	cs.params[vn], err = convValue(q.Val, q.Type)
	return result, err
}

func (cs *compilerSession) compileOperatorBETWEEN(q *AtomaryCondition) (string, error) {
	rng, ok := q.Val.([]any)
	if !ok {
		return "", errors.New("w3sql: wrong value for field " + q.Col)
	}
	v, err := convRange(rng, q.Type)
	if err != nil {
		return "", err
	}

	vn1 := "sqv" + fmt.Sprint(cs.varCounter) + "_1"
	vn2 := "sqv" + fmt.Sprint(cs.varCounter) + "_2"
	cs.varCounter++

	result := ""
	if field, ok := cs.getSearchField(q.Col, q.Type); ok {
		result = fmt.Sprintf("(%v>=:%v AND %v<=:%v)", field, vn1, field, vn2)
	} else {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}
	cs.params[vn1] = v.from
	cs.params[vn2] = v.to

	return result, nil
}

func (cs *compilerSession) compileOperatorReverseIN(q *AtomaryCondition) (string, error) {
	var (
		err   error
		field string
		ok    bool
	)
	if field, ok = cs.getSearchField(q.Col, q.Type); !ok {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}
	vn := fmt.Sprintf("sqv%d_1", cs.varCounter)
	cs.varCounter++
	cs.params[vn], err = convValueElem(q.Val, q.Type)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(":%v in (%v)", vn, field), nil
}

func (cs *compilerSession) compileOperatorIN(q *AtomaryCondition, not bool) (string, error) {
	var (
		err   error
		field string
		ok    bool
		rng   []any
	)
	if field, ok = cs.getSearchField(q.Col, q.Type); !ok {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}

	if rng, ok = q.Val.([]any); !ok {
		return "", errors.New("w3sql: wrong field value " + q.Col)
	}
	searchStr := "("
	if not {
		searchStr += "not "
	}
	searchStr += fmt.Sprintf("%v in (", field)
	for j, cv := range rng {
		vn := fmt.Sprintf("sqv%d_%d", cs.varCounter, j)
		cs.params[vn], err = convValueElem(cv, q.Type)
		if err != nil {
			return "", err
		}
		searchStr += fmt.Sprintf(":%v,", vn)
	}
	cs.varCounter++
	searchStr = strings.TrimRight(searchStr, ",") + "))"
	return searchStr, nil
}

func (cs *compilerSession) compileOperatorBEGINS(q *AtomaryCondition, contains, ends bool) (string, error) {
	vn := "sqv" + fmt.Sprint(cs.varCounter)
	cs.varCounter++
	result := ""
	var err error
	if field, ok := cs.getSearchField(q.Col, q.Type); ok {
		op := "(%v LIKE :%v || '%%')"
		if contains {
			op = "(%v LIKE '%%' || :%v || '%%')"
		} else if ends {
			op = "(%v LIKE '%%' || :%v)"
		}
		result = fmt.Sprintf(op, field, vn)
	} else {
		return "", errors.New("w3sql: no such field name " + q.Col)
	}

	cs.params[vn], err = convValue(q.Val, q.Type)
	if err != nil {
		return "", err
	}
	return result, nil
}
