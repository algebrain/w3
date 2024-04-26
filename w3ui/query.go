package w3ui

import (
	"encoding/json"

	"github.com/algebrain/w3/w3sql"
)

func FromJSON(s string) (*Query, error) {
	var q_ w3sql.Query
	err := json.Unmarshal([]byte(s), &q_)
	if err != nil {
		return nil, err
	}

	q := Query(q_)
	return &q, nil
}
