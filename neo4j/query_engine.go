package neo4j

import (
	lucy "Neo4j-OGM/core"
	lucyErr "Neo4j-OGM/errors"
)

type CypherCradle struct {
	exps []lucy.Expr
	ops  []string
}

func (c *CypherCradle) init() {
	c.exps = make([]lucy.Expr, 0)
	c.ops = make([]string, 0)
}

type QueryEngine struct {
	queue      *lucy.QueryQueue
	endDomain  lucy.DomainType
	hasStarted bool
	cradle     *CypherCradle
}

func (q *QueryEngine) NewQueryEngine() lucy.Layer {
	q.endDomain = lucy.Unknown
	q.cradle = &CypherCradle{}
	q.cradle.init()

	return q
}

func (q *QueryEngine) AttachTo(l *lucy.Lucy) {
	q.queue = &l.Queue
	l.SetLayer(q)
}

func (q *QueryEngine) Start() {
	q.hasStarted = true
}

func (q *QueryEngine) Sync() error {
	if !q.hasStarted {
		qr, err := q.queue.Get()
		if err != nil {
			return err
		}

		switch qr.DomainType {
		case lucy.Where:
			{
				q.cradle.exps = append(q.cradle.exps, qr.Params.(lucy.Expr))
			}
		}

		return nil
	}

	for !q.queue.IsEmpty() {
		qr, err := q.queue.Get()
		if err != nil {
			return err
		}

		// Manage end-domain for the query engine.
		if lucy.IsEndDomain(qr.DomainType) {
			if q.endDomain != lucy.Unknown {
				if q.endDomain != qr.DomainType {
					return lucyErr.EndDomainChanged
				}
			} else {
				q.endDomain = qr.DomainType
			}
		}

		switch qr.DomainType {

		}

	}

	return nil
}

func (q *QueryEngine) Stop() {
	q.hasStarted = false
}
