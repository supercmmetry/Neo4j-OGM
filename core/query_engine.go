package lucy

import (
	lucyErr "Neo4j-OGM/errors"
)

type QueryCradle struct {
	exps      []Expr
	ops       []string
	dom, pdom DomainType
	deps      map[string]struct{}
	out       interface{}
}

func (c *QueryCradle) init() {
	c.dom = Unknown
	c.pdom = Unknown
	c.exps = make([]Expr, 0)
	c.ops = make([]string, 0)
	c.deps = make(map[string]struct{})

}

type QueryRuntime interface {
	Compile(cradle *QueryCradle) (string, error)
	Execute(query string, target interface{}) error
}

type QueryEngine struct {
	queue         *QueryQueue
	hasStarted    bool
	isTransaction bool
	cradle        *QueryCradle
	Runtime       QueryRuntime
}

func (q *QueryEngine) AddRuntime(rt QueryRuntime) {
	q.Runtime = rt
}

func (q *QueryEngine) NewQueryEngine() Layer {
	q.cradle = &QueryCradle{}
	q.cradle.init()
	q.isTransaction = false
	return q
}

func (q *QueryEngine) AttachTo(l *Database) {
	q.queue = &l.Queue
	l.SetLayer(q)
}

func (q *QueryEngine) Start() {
	if q.isTransaction {
		return
	}
	q.hasStarted = true
}

func (q *QueryEngine) StartTransaction() {
	q.isTransaction = true
}

func (q *QueryEngine) Sync() error {
	if !q.hasStarted {
		qr, err := q.queue.Get()
		if err != nil {
			return err
		}

		switch qr.DomainType {
		case Where:
			{
				if q.cradle.pdom == Where {
					return lucyErr.QueryChainLogicCorrupted
				}
				q.cradle.exps = append(q.cradle.exps, qr.Params.(Expr))
				q.cradle.deps["where"] = struct{}{}
			}
		case And:
			{
				if _, ok := q.cradle.deps["where"]; !ok {
					return lucyErr.QueryDependencyNotSatisfied
				}
				q.cradle.exps = append(q.cradle.exps, qr.Params.(Expr))
				q.cradle.ops = append(q.cradle.ops, "and")
			}
		case Or:
			{
				if _, ok := q.cradle.deps["where"]; !ok {
					return lucyErr.QueryDependencyNotSatisfied
				}
				q.cradle.exps = append(q.cradle.exps, qr.Params.(Expr))
				q.cradle.ops = append(q.cradle.ops, "or")
			}
		}

		q.cradle.pdom = q.cradle.dom
		return nil
	}

	for !q.queue.IsEmpty() {
		qr, err := q.queue.Get()
		if err != nil {
			return err
		}

		// Manage end-domain for the query engine.
		if IsEndDomain(qr.DomainType) {
			if q.cradle.dom != Unknown {
				if q.cradle.dom != qr.DomainType {
					return lucyErr.EndDomainChanged
				}
			} else {
				q.cradle.dom = qr.DomainType
			}
		}

		switch qr.DomainType {
		case SetTarget:
			{
				q.cradle.out = qr.Output
				// todo: Evaluate query with query-compiler.
			}
		}

	}

	return nil
}

func (q *QueryEngine) Stop() {
	q.hasStarted = false
}
