package w3sql

import (
	"encoding/json"
	"fmt"
	"testing"
)

var insertJSON = `{
	"Insert": {
		"Cols": ["name", "age", "score"],
		"Values": [
			["Vanya", 21, 90],
			["Masha", 20, 91],
			["Petya", 19, 92]
		]
	}
}`

var insertBaseSQL = NewSQLString("insert into students")

func TestCompileInsert(t *testing.T) {
	var q Query
	err := json.Unmarshal([]byte(insertJSON), &q)
	if err != nil {
		t.Fatal(err)
	}

	transformScore := func(field string, value any) (any, error) {
		if field == "name" && value.(string) == "Masha" {
			return nil, nil
		}
		if field != "score" {
			return value, nil
		}
		return value.(float64) / 100, nil
	}

	iq, err := q.CompileInsert("sqlite", map[string]string{
		"name":  "",
		"age":   "",
		"score": "score_value",
	}, transformScore)

	if err != nil {
		t.Fatal(err)
	}

	qs, err := iq.SQL(insertBaseSQL)
	if err != nil {
		t.Fatal(err)
	}
	p := qs[0].Params

	if len(p) != len(iq.Cols)*len(iq.Values) {
		t.Fatal("wrong length of parameters table, got", len(p), "expected", len(iq.Cols)*len(iq.Values))
	}

	fmt.Println("QUERY:", qs[0].Code)
	fmt.Println("PARAMS:", p)

	expectedQS := `
insert into students (name,age,score_value)
values
(:uiname0,:uiage1,:uiscore2),
(:uiname3,:uiage4,:uiscore5)`
	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}
}

var updateJSON = `{
	"Update": {
		"Cols": ["name", "age", "score", "id"],
		"Values": [
			["Vanya", 21, 90, 1],
			["Masha", 20, 91, 2],
			["Petya", 19, 92, 3]
		]
	}
}`

var updateBaseSQL = NewSQLString("update students")

func TestCompileUpdate(t *testing.T) {
	var q Query
	err := json.Unmarshal([]byte(updateJSON), &q)
	if err != nil {
		t.Fatal(err)
	}

	transformScore := func(field string, value any) (any, error) {
		if field == "name" && value.(string) == "Masha" {
			return nil, nil
		}
		if field != "score" {
			return value, nil
		}
		return value.(float64) / 100, nil
	}

	uq, err := q.CompileUpdate("sqlite", map[string]string{
		"name":  "",
		"age":   "",
		"score": "score_value",
		"id":    "",
	}, "id", transformScore)

	if err != nil {
		t.Fatal(err)
	}

	qs, err := uq.SQL(updateBaseSQL)
	if err != nil {
		t.Fatal(err)
	}
	p := qs[0].Params

	if len(p) != len(uq.Cols)*len(uq.Values) {
		t.Fatal("wrong length of parameters table, got", len(p), "expected", len(uq.Cols)*len(uq.Values))
	}

	fmt.Println("QUERY:", qs[0].Code)
	fmt.Println("PARAMS:", p)
	expectedQS := `
update students set
name = c.name,
age = c.age,
score_value = c.score_value,
from (values
(:uiname0,:uiage1,:uiscore2,:uiid3),
(:uiname4,:uiage5,:uiscore6,:uiid7)
) as c(name,age,score_value,id)
where id = c.id`
	if !EqualSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0]),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}
}
