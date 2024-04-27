package w3sql

import (
	"encoding/json"
	"fmt"
	"testing"
)

var atomaryJSON = `{
	"Offset": 20,
	"Limit": 10,
	"Sort": [{"Col": "name", "Dir": "desc"}],
	"Search": {
		"Col": "age", 
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
	qs, err := cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0].Code)
	fmt.Println("Params:", qs[0].Params)

	expectedQS := `
select * from students
where (age::int<=:sqv0)
limit 10
offset 20
order by name DESC`

	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	base = NewSQLString("select * from students where score > 50")
	qs, err = cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}

	fmt.Println("Query:", qs[0].Code)
	fmt.Println("Params:", qs[0].Params)
	p := qs[0].Params

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

	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}
}

var compoundJSON = `{
	"Offset": 20,
	"Limit": 10,
	"Sort": [{"Col": "name", "Dir": "desc"}],
	"Search": {
		"Op": "and",
		"Query": [
			{"Col": "age", "Type": "int", "Val": 23, "Op": "<="},
			{
				"Op": "or",
				"Query": [
					{
						"Col": "name",
						"Type":  "string",
						"Val": "Bob",
						"Op":    "contains"
					},
					{
						"Col": "name",
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
	qs, err := cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}

	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len(qs))
	}

	p := qs[0].Params
	if len(p) != 3 {
		t.Fatal("3 parameters expected, got", len(p))
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

	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("\n<%s>", NormalizeSQLString(qs[0].Code, true)),
			"\nexpected",
			fmt.Sprintf("\n<%s>", NormalizeSQLString(expectedQS, true)),
		)
	}

	base = NewSQLString("select * from students where score > 50")
	qs, err = cq.SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	p = qs[0].Params
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0].Code)
	fmt.Println("Params:", p)

	expectedQS = `
select * from students where score > 50
and ((age::int<=:sqv0) AND ((name LIKE '%' || :sqv1 || '%') OR (name LIKE :sqv2 || '%')))
limit 10
offset 20
order by name DESC`

	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
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
	qs, err := cq.NoConditions().SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	p := qs[0].Params
	fmt.Println("Query:", qs[0].Code)
	fmt.Println("Params:", p)

	expectedQS := `select * from students
limit 10
offset 20
order by name DESC`
	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	qs, err = cq.NoLimitOffset().SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	p = qs[0].Params
	fmt.Println("Query:", qs[0].Code)
	fmt.Println("Params:", p)
	expectedQS = `select * from students
where (age::int<=:sqv0)
order by name DESC`
	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	qs, err = cq.NoConditions().SQL(base)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	p = qs[0].Params
	fmt.Println("Query:", qs[0].Code)
	fmt.Println("Params:", p)
	expectedQS = `select * from students
limit 10
offset 20
order by name DESC`
	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

}
