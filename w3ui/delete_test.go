package w3ui

import (
	"testing"
)

func TestCompileDelete(t *testing.T) {
	tableIds := map[string]string{
		"events": "id",
	}

	s := `{
		"Delete": [4]
	}`

	q, err := FromJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	fd := func(id string) (bool, string) { return true, "" }

	sqs, serr := q.CompileDeleteSql(tableIds, fd)
	if serr != "" {
		t.Fatal(serr)
	}
	if len(sqs) != 1 {
		t.Fatal("wrong []SqlQuery length")
	}
	if sqs[0].Text == "" {
		t.Fatal("empty text")
	}
	if sqs[0].Text != "delete from events where id = :ui0" {
		t.Fatalf("incorrect Text: |%s|", sqs[0].Text)
	}
	if len(sqs[0].Map) == 0 {
		t.Fatal("empty params map")
	}

	v, ok := sqs[0].Map["ui0"]
	if !ok {
		t.Fatal("bad map values")
	}
	if v.(float64) != 4 {
		t.Fatalf("incorrect map value: %v", v)
	}

	s = `{
		"Delete": [4, 6]
	}`

	q, err = FromJSON(s)
	if err != nil {
		t.Fatal(err)
	}

	sqs, serr = q.CompileDeleteSql(tableIds, fd)
	if serr != "" {
		t.Fatal(serr)
	}
	if len(sqs) != 1 {
		t.Fatal("wrong []SqlQuery length 2")
	}
	if sqs[0].Text == "" {
		t.Fatal("empty text 2")
	}
	if sqs[0].Text != "delete from events where id in (:ui0,:ui1)" {
		t.Fatalf("incorrect Text 2: |%s|", sqs[0].Text)
	}
	if len(sqs[0].Map) != 2 {
		t.Fatal("wrong map params count")
	}

	v, ok = sqs[0].Map["ui0"]
	if !ok {
		t.Fatal("bad map values 2")
	}
	if v.(float64) != 4 {
		t.Fatalf("incorrect map value 2: %v", v)
	}
	v, ok = sqs[0].Map["ui1"]
	if !ok {
		t.Fatal("bad map values 3")
	}
	if v.(float64) != 6 {
		t.Fatalf("incorrect map value 3: %v", v)
	}

	tableIds["logs"] = "logid"
	sqs, serr = q.CompileDeleteSql(tableIds, fd)
	if serr != "" {
		t.Fatal(err)
	}
	if len(sqs) != 2 {
		t.Fatal("wrong []SqlQuery length 3")
	}
	if sqs[0].Text == "" {
		t.Fatal("empty text 3")
	}
	if sqs[1].Text == "" {
		t.Fatal("empty text 4")
	}
	if sqs[0].Text != "delete from events where id in (:ui0,:ui1)" &&
		sqs[0].Text != "delete from logs where logid in (:ui0,:ui1)" {
		t.Fatalf("incorrect Text 3: |%s|", sqs[0].Text)
	}
	if len(sqs[0].Map) != 2 {
		t.Fatal("wrong map params count 3")
	}
	if sqs[1].Text != "delete from events where id in (:ui0,:ui1)" &&
		sqs[1].Text != "delete from logs where logid in (:ui0,:ui1)" {
		t.Fatalf("incorrect Text 3: |%s|", sqs[1].Text)
	}
	if len(sqs[1].Map) != 2 {
		t.Fatal("wrong map params count 3")
	}
	if sqs[0].Text == sqs[1].Text {
		t.Fatal("both sql texts are equal")
	}
}
