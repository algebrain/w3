package w3sql

import (
	"encoding/json"
	"fmt"
	"testing"
)

var deleteJSON = `{
	"Delete": [1,2,3,4,5]
}`

func TestCompileDelete(t *testing.T) {
	var q Query
	err := json.Unmarshal([]byte(deleteJSON), &q)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("UNMARSHALLED QUERY:", q)

	transformDel := func(tableName string, idName string, id any) (any, error) {
		if tableName == "students" && id.(float64) == 3 {
			return nil, nil //prohibit deletion
		}
		if tableName == "avatars" {
			return fmt.Sprint(id), nil
		}
		return id, nil
	}

	dq, err := q.CompileDelete("sqlite", []*TableDelete{
		&TableDelete{
			TableName: "students",
			IDName:    "studentID",
		},
		&TableDelete{
			TableName: "avatars",
			IDName:    "imageID",
		},
	}, transformDel)

	if err != nil {
		t.Fatal(err)
	}

	b, _ := json.MarshalIndent(dq, "", "  ")
	fmt.Println("COMPILED DELETE:", string(b))
	qs, err := dq.SQL(nil)
	if err != nil {
		t.Fatal(err)
	}

	for _, qq := range qs {
		fmt.Println("QUERY:", qq.Code)
		fmt.Println("PARAMS:", qq.Params)
	}

	if len(qs) != 2 {
		t.Fatal("2 sql queries expected, got ", len(qs))
	}
	expectedQS := "delete from students where studentID in (:ui0,:ui1,:ui2,:ui3)"
	if !equalSQLStrings(expectedQS, qs[0].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[0].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	expectedQS = "delete from avatars where imageID in (:ui0,:ui1,:ui2,:ui3,:ui4)"
	if !equalSQLStrings(expectedQS, qs[1].Code) {
		t.Fatal(
			"unexpected sql string result, got:",
			fmt.Sprintf("<%s>", qs[1].Code),
			"\nexpected",
			fmt.Sprintf("<%s>", expectedQS),
		)
	}

	if len(qs[0].Params) != 4 {
		t.Fatal("4 parameters expected, got", qs[0].Params)
	}

	if len(qs[1].Params) != 5 {
		t.Fatal("5 parameters expected, got", qs[1].Params)
	}
}
