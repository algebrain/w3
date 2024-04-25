package w3sql

import (
	"encoding/json"
	"testing"
)

func TestJSON(t *testing.T) {
	s := `{
		"Offset": 10,
		"Limit": 20,
		"Sort": [{"Field": "name", "Direction": "desc"}],
		"Search": {
			"Op": "and",
			"Query": [
				{"Field": "age", "Type": "int", "Value": 23, "Op": "<="},
				{
					"Op": "or",
					"Query": [
						{
							"Field": "name",
							"Type":  "string",
							"Value": "Bob",
							"Op":    "contains"
						},
						{
							"Field": "name",
							"Type":  "string",
							"Value": "Alice",
							"Op":    "starts with"
						}
					]
				}
			]
		}
	}`

	var q Query
	err := json.Unmarshal([]byte(s), &q)
	if err != nil {
		t.Fatal(err)
	}

	if q.Offset == nil || *q.Offset != 10 {
		t.Fatal("wrong 'Offset'", q.Offset)
	}
	if q.Limit == nil || *q.Limit != 20 {
		t.Fatal("wrong 'Limit'", q.Limit)
	}
	if q.Search == nil {
		t.Fatal("'Search' is nil")
	}

	if qq, ok := q.Search.(*CompoundCondition); !ok {
		t.Fatal("'CompoundCondition' expected for 'Search'")
	} else if qq.Op != "and" {
		t.Fatal("'and' expected for 'Op' in 'Search'")
	}
}
