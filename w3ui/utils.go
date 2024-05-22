package w3ui

import (
	"encoding/json"
	"strings"
)

func joinNonEmpty(s []string, delim string) string {
	ss := make([]string, 0, len(s))
	for _, b := range s {
		if b != "" {
			ss = append(ss, b)
		}
	}

	return strings.Join(ss, delim)
}

func GetJSON(x any) string {
	b, err := json.MarshalIndent(&x, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(b)
}
