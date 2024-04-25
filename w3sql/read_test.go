package w3sql

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"
)

var atomaryJSON = `{
	"Offset": 20,
	"Limit": 10,
	"Sort": [{"Field": "name", "Dir": "desc"}],
	"Search": {
		"Field": "age", 
		"Type": "int", 
		"Val": 23, 
		"Op": "<="
	}
}`

func TestCompileAtomarySelect(t *testing.T) {
	s := atomaryJSON
	var q Query
	err := json.Unmarshal([]byte(s), &q)
	if err != nil {
		t.Fatal(err)
	}

	cq, err := q.CompileSelect("sqlite", map[string]string{"age": "age::int", "name": ""})
	if err != nil {
		t.Fatal(err)
	}

	base := NewSQLString("select * from students")
	qs, p, err := cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)

	expectedQS := `
select * from students
where (age::int<=:sqv0)
limit 10
offset 20
order by name DESC`

	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	base = NewSQLString("select * from students where score > 50")
	qs, p, err = cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)

	if len(p) != 1 {
		t.Fatal("1 parameter expected, got", len((p)))
	}

	if x, ok := p["sqv0"].(float64); !ok || x != 23 {
		t.Fatal("expected 23 for age, got", fmt.Sprintf("<%v> <%T> <%v>", x, p["sqv0"], ok))
	}

	expectedQS = `
select * from students where score > 50
and (age::int<=:sqv0)
limit 10
offset 20
order by name DESC`

	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}
}

var compoundJSON = `{
	"Offset": 20,
	"Limit": 10,
	"Sort": [{"Field": "name", "Dir": "desc"}],
	"Search": {
		"Op": "and",
		"Query": [
			{"Field": "age", "Type": "int", "Val": 23, "Op": "<="},
			{
				"Op": "or",
				"Query": [
					{
						"Field": "name",
						"Type":  "string",
						"Val": "Bob",
						"Op":    "contains"
					},
					{
						"Field": "name",
						"Type":  "string",
						"Val": "Alice",
						"Op":    "starts with"
					}
				]
			}
		]
	}
}`

func TestCompileCompoundSelect(t *testing.T) {
	s := compoundJSON
	var q Query
	err := json.Unmarshal([]byte(s), &q)
	if err != nil {
		t.Fatal(err)
	}

	cq, err := q.CompileSelect("sqlite", map[string]string{"age": "age::int", "name": ""})
	if err != nil {
		t.Fatal(err)
	}

	base := NewSQLString("select * from students")
	qs, p, err := cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}

	if len(p) != 3 {
		t.Fatal("3 parameters expected, got", len((p)))
	}

	values := map[string]int{"23": 1, "Bob": 1, "Alice": 1}
	for _, v := range p {
		vv := fmt.Sprint(v)
		if _, ok := values[vv]; !ok {
			t.Fatal("no such value expected", v)
		}
		delete(values, vv)
	}
	if len(values) != 0 {
		t.Fatal("not all expected values are received: ", values)
	}

	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)

	expectedQS := `
select * from students
where ((age::int<=:sqv0) AND ((name LIKE '%' || :sqv1 || '%') OR (name LIKE :sqv2 || '%')))
limit 10
offset 20
order by name DESC`

	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("\n<%s>", normalizeSQLString(strings.TrimSpace(qs[0]))),
			"\nexpected",
			fmt.Sprintf("\n<%s>", normalizeSQLString(strings.TrimSpace(expectedQS))),
		)
	}

	base = NewSQLString("select * from students where score > 50")
	qs, p, err = cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)

	expectedQS = `
select * from students where score > 50
and ((age::int<=:sqv0) AND ((name LIKE '%' || :sqv1 || '%') OR (name LIKE :sqv2 || '%')))
limit 10
offset 20
order by name DESC`

	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}
}

func TestCompileNo(t *testing.T) {
	s := atomaryJSON
	var q Query
	err := json.Unmarshal([]byte(s), &q)
	if err != nil {
		t.Fatal(err)
	}

	cq, err := q.CompileSelect("sqlite", map[string]string{"age": "age::int", "name": ""})
	if err != nil {
		t.Fatal(err)
	}

	base := NewSQLString("select * from students")
	qs, p, err := cq.NoConditions().SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)

	expectedQS := `select * from students
limit 10
offset 20
order by name DESC`
	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	qs, p, err = cq.NoLimitOffset().SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)
	expectedQS = `select * from students
where (age::int<=:sqv0)
order by name DESC`
	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	qs, p, err = cq.NoConditions().SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	fmt.Println("Params:", p)
	expectedQS = `select * from students
limit 10
offset 20
order by name DESC`
	if !equalSQLStrings(expectedQS, qs[0]) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

}
