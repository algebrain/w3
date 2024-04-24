package w3sql

import (
	"fmt"
	"testing"
)

func TestCompileAtomarySelect(t *testing.T) {
	limit := 10
	offset := 20
	q := Query{
		Limit:  &limit,
		Offset: &offset,
		Search: &AtomaryCondition{
			Field: "age",
			Type:  "int",
			Value: 23,
			Op:    "<=",
		},
		Sort: []SortQuery{
			SortQuery{
				Field:     "name",
				Direction: "desc",
			},
		},
	}

	cq, err := q.CompileSelect("sqlite", map[string]string{"age": "age::int", "name": ""})
	if err != nil {
		t.Fatal(err)
	}

	base := "select * from students"
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

	base = "select * from students where score > 50"
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

func TestCompileCompoundSelect(t *testing.T) {
	limit := 10
	offset := 20
	q := Query{
		Limit:  &limit,
		Offset: &offset,
		Search: &CompoundCondition{
			Op: "and",
			Query: []RawCondition{
				&AtomaryCondition{
					Field: "age",
					Type:  "int",
					Value: 23,
					Op:    "<=",
				},
				&CompoundCondition{
					Op: "or",
					Query: []RawCondition{
						&AtomaryCondition{
							Field: "name",
							Type:  "string",
							Value: "Bob",
							Op:    "contains",
						},
						&AtomaryCondition{
							Field: "name",
							Type:  "string",
							Value: "Alice",
							Op:    "starts with",
						},
					},
				},
			},
		},
		Sort: []SortQuery{
			SortQuery{
				Field:     "name",
				Direction: "desc",
			},
		},
	}

	cq, err := q.CompileSelect("sqlite", map[string]string{"age": "age::int", "name": ""})
	if err != nil {
		t.Fatal(err)
	}

	base := "select * from students"
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
where ((age::int<=:sqv0) AND ((name LIKE '%' || :sqv1 || '%') OR (name LIKE :sqv2 || '%')))
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

	base = "select * from students where score > 50"
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
