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
		fmt.Println("QUERY:", qq.code)
		fmt.Println("PARAMS:", qq.params)
	}
}
