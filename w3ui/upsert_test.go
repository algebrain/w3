package w3ui

import (
	"testing"
	"time"
)

func TestCompileUpdate(t *testing.T) {
	s := `{
		"Update": {
			"Cols": ["name", "email", "id"],
			"Values": [
				["test", "test@gmail.com", 1],
				["test1", "test@gmail.com", 0]
			]
		},
		"Insert": {
			"Cols": ["name", "email", "created"],
			"Values": [
				["test3", "test3@gmail.com", 0],
				["test4", "test4@gmail.com", 0]
			]
		}
	}`

	q, err := ReadJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	fieldMap := map[string]bool{
		"name":    true,
		"email":   true,
		"id":      true,
		"created": true,
	}

	beforeCreate := func(isInsert bool, f string, v interface{}) (interface{}, error) {
		if f == "created" {
			if isInsert {
				return time.Now().Unix(), nil
			} else {
				return nil, nil
			}
		}
		return v, nil
	}

	sq, err := q.CompileUpsert("id", fieldMap, beforeCreate)
	t.Log(getJSON(sq))

	if err != nil {
		t.Fatal(err)
	}
	if len(sq.UpdateQueries) == 0 {
		t.Fatal("bad update queries")
	}

	if len(sq.InsertQueries) == 0 {
		t.Fatal("bad insert query")
	}
	if v, ok := sq.UpdateValues["uiid2"]; !ok || v == "" {
		t.Fatal("bad update values")
	}
	if v, ok := sq.UpdateValues["uiname0"]; !ok || v == "" {
		t.Fatal("bad update values")
	}

	if len(sq.InsertQueries) == 0 {
		t.Fatal("bad insert query")
	}
	if v, ok := sq.InsertValues["uiname0"]; !ok || v == "" {
		t.Fatal("bad insert values")
	}

	if v, ok := sq.InsertValues["uicreated2"]; !ok || v == "" {
		t.Fatal("bad insert values")
	}
}
