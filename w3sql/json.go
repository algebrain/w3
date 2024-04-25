package w3sql

import (
	"encoding/json"
)

type jsonCondition struct {
	Field string
	Type  string
	Val   any
	Op    string
	Query []*jsonCondition
}

type jsonQuery struct {
	Limit  *int
	Offset *int
	Search *jsonCondition
	Sort   []SortQuery
	Insert *struct {
		Fields []string
		Values [][]any
	}
	Update *struct {
		Fields []string
		Values [][]any
	}
	Delete []Record
	Params map[string]any //дополнительные параметры запроса, вне логики SQL
}

func (c *jsonCondition) read() RawCondition {
	if c.Query != nil {
		q := make([]RawCondition, len(c.Query))
		for i, s := range c.Query {
			q[i] = s.read()
		}
		return &CompoundCondition{
			Query: q,
			Op:    c.Op,
		}
	}
	return &AtomaryCondition{
		Field: c.Field,
		Type:  c.Type,
		Val:   c.Val,
		Op:    c.Op,
	}
}

func (q *Query) UnmarshalJSON(data []byte) error {
	var raw jsonQuery
	err := json.Unmarshal(data, &raw)
	if err != nil {
		return err
	}
	q.Limit = raw.Limit
	q.Offset = raw.Offset
	if raw.Search != nil {
		q.Search = raw.Search.read()
	}
	q.Sort = raw.Sort
	q.Insert = raw.Insert
	q.Update = raw.Update
	q.Delete = raw.Delete
	q.Params = raw.Params
	return nil
}
