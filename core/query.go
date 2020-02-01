package lucy

import lucyErr "Neo4j-OGM/errors"

type QueryQueue struct {
	queries *[]Query
}

func (q *QueryQueue) Init() {
	queries := make([]Query, 0)
	q.queries = &queries
}

func (q *QueryQueue) Push(query Query) {
	*q.queries = append(*q.queries, query)
}

func (q *QueryQueue) Get() (Query, error) {
	if (len(*q.queries)) == 0 {
		return Query{}, lucyErr.EmptyQueryQueue
	}
	query := (*q.queries)[0]
	*q.queries = (*q.queries)[1:]
	return query, nil
}

func (q *QueryQueue) IsEmpty() bool {
	return len(*q.queries) == 0
}

type Query struct {
	DomainType DomainType
	Params     interface{}
	Output     interface{}
}
