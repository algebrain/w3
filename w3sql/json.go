package w3sql

import (
	"encoding/json"
)

type jsonCondition struct {
	Field string
	Type  string
	Value any
	Op    string
	Query []*jsonCondition
}

type jsonQuery struct {
	Limit  *int
	Offset *int
	Search *jsonCondition
	Sort   []SortQuery
	Insert []Record
	Update []Record
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
		Value: c.Value,
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
	q.Search = raw.Search.read()
	q.Sort = raw.Sort
	q.Insert = raw.Insert
	q.Update = raw.Update
	q.Delete = raw.Delete
	q.Params = raw.Params
	return nil
}
