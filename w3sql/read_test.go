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

func TestCompileNeedWherePrepared(t *testing.T) {
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

	base := ""
	qs, _, err := cq.SQL(base, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	expectedQS := `
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

	qs, _, err = cq.SQL(base, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(qs) != 1 {
		t.Fatal("1 sql query expected, got", len((qs)))
	}
	fmt.Println("Query:", qs[0])
	expectedQS = `
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
