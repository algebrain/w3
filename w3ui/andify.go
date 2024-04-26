package w3ui

import (
	"strings"

	"github.com/algebrain/w3/w3sql"
)

func AndifyReq(before string, req string) string {
	if w3sql.NeedsWhere(before) {
		return " " + req
	}
	return strings.Replace(req, "where", "and", 1)
}
