package w3ui

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/algebrain/w3/w3sql"
)

func TestQuery(t *testing.T) {
	s := `{
		"Limit": 50,
		"Sort": [
			{"Field": "fname", "Dir": "asc"},
			{"Field": "lname", "Dir": "desc"}
		],
		"Search": {
			"Op": "AND",
			"Query": [
				{
					"Field": "fname",
					"Type": "text",
					"Op": "равен",
					"Val": "vit"
				},
				{
					"Field": "age",
					"Type": "number",
					"Op": "между",
					"Val": [10, 20]
				}
			]
		}
	}`

	q, err := FromJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("rcq: %+v", q)

	qs, err := q.Compile(map[string]string{
		"fname": "",
		"age":   "",
		"lname": "",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	b, _ := json.MarshalIndent(qs.Map, "", "  ")
	t.Log(qs.Text)
	t.Log(string(b))

	good := "where ((fname=:sqv0) AND (age>=:sqv1_1 AND age<=:sqv1_2))  order by fname ASC, lname DESC limit 50"
	if !w3sql.EqualSQLStrings(qs.Text, good) {
		have := w3sql.NormalizeSQLString(strings.TrimSpace(qs.Text))
		want := w3sql.NormalizeSQLString(strings.TrimSpace(good))
		t.Fatal("\n"+have, "\n"+want)
	}
}

func TestQuery1(t *testing.T) {
	s := `{
		"Limit": 50,
		"Sort": [
			{"Field": "fname", "Dir": "asc"},
			{"Field": "lname", "Dir": "desc"}
		],
		"Search": {
			"Op": "OR",
			"Query": [
				{
					"Field": "fname",
					"Type": "text",
					"Op": "равен",
					"Val": "vit"
				},
				{
					"Field": "age",
					"Type": "number",
					"Op": "между",
					"Val": [10, 20]
				}
			]
		}
	}`

	q, err := FromJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("rcq: %+v", q)

	qs, err := q.Compile(map[string]string{
		"fname": "fname",
		"age":   "age",
		"lname": "lname",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	b, _ := json.MarshalIndent(qs.Map, "", "  ")
	t.Log(qs.Text)
	t.Log(string(b))

	good := "where ((fname=:sqv0) OR (age>=:sqv1_1 AND age<=:sqv1_2))  order by fname ASC, lname DESC limit 50"
	if !w3sql.EqualSQLStrings(qs.Text, good) {
		have := w3sql.NormalizeSQLString(strings.TrimSpace(qs.Text))
		want := w3sql.NormalizeSQLString(strings.TrimSpace(good))
		t.Fatal("\n"+have, "\n"+want)
	}
}

func TestQuery3(t *testing.T) {
	s := `{
		"Limit": 50,
		"Search": {
			"Op": "OR",
			"Query": [
				{
					"Field": "UserName",
					"Type": "text",
					"Op": "contains",
					"Val": "ame"
				},
				{
					"Field": "ChannelName",
					"Type": "text",
					"Op": "contains",
					"Val": "ame"
				}
			]
		}
	}`

	q, err := FromJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("rcq: %+v", *q)

	qs, err := q.Compile(map[string]string{
		"UserName":    "fname",
		"age":         "age",
		"ChannelName": "lname",
	})

	if err != nil {
		t.Fatal(err.Error())
	}

	b, _ := json.MarshalIndent(qs.Map, "", "  ")
	t.Log(qs.Text)
	t.Log(string(b))

	good := "where ((fname LIKE '%' || :sqv0 || '%') OR (lname LIKE '%' || :sqv1 || '%'))  limit 50"
	if !w3sql.EqualSQLStrings(qs.Text, good) {
		have := w3sql.NormalizeSQLString(strings.TrimSpace(qs.Text))
		want := w3sql.NormalizeSQLString(strings.TrimSpace(good))
		t.Fatal("\n"+have, "\n"+want)
	}
}
