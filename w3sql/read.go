package w3sql

type SelectQuery struct {
	CompiledQueryParams
	Conditions string //логические ограничения, например age < 35 and name='John'
	Limit      *int
	Offset     *int
	Order      []string //например age desc
}

func (q *Query) CompileSelect(
	sqlSyntax string,
	fieldmap map[string]string,
) (*SelectQuery, error) {
	result := &SelectQuery{
		Limit:  q.Limit,
		Offset: q.Offset,
		CompiledQueryParams: CompiledQueryParams{
			Params: q.Params,
		},
	}
	cs := &compilerSession{
		sqlSyntax: sqlSyntax,
		fieldmap:  fieldmap,
		params:    map[string]any{},
	}
	var err error
	if q.Search != nil {
		result.Conditions, err = q.Search.compile(cs)
		if err != nil {
			return nil, err
		}
	}
	if q.Sort != nil && len(q.Sort) > 0 {
		result.Order = make([]string, len(q.Sort))
		for i, sq := range q.Sort {
			p, err := sq.compile(cs)
			if err != nil {
				return nil, err
			}
			result.Order[i] = p
		}
	}
	result.CompiledQueryParams.SQLParams = cs.params
	return result, nil
}
