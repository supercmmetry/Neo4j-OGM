package lucy

import (
	lucyErr "lucy/errors"
)

type QueryCradle struct {
	exps      Queue
	ops       Queue
	dom, pdom DomainType
	deps      map[DomainType]struct{}
	out       interface{}
}

func (c *QueryCradle) init() {
	c.dom = Unknown
	c.pdom = Unknown
	c.exps.Init()
	c.ops.Init()
	c.deps = make(map[DomainType]struct{})
}

type QueryRuntime interface {
	Compile(cradle *QueryCradle) (string, error)
	Execute(query string, target interface{}) error
}

type QueryEngine struct {
	queue         *Queue
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

func (q *QueryEngine) StartTransaction() {
	q.isTransaction = true
}

func (q *QueryEngine) Sync() error {
	cradle := q.cradle
	for !q.queue.IsEmpty() {
		qri, err := q.queue.Get()
		if err != nil {
			return err
		}

		qr := qri.(Query)

		cradle.dom = qr.DomainType
		switch qr.DomainType {
		case Where:
			{
				if cradle.pdom == Where {
					return lucyErr.QueryChainLogicCorrupted
				}
				cradle.exps.Push(qr.Params.(Exp))
				cradle.ops.Push(Where)

				cradle.deps[Where] = struct{}{}
			}
		case WhereStr: {
			if cradle.pdom == Where {
				return lucyErr.QueryChainLogicCorrupted
			}
			cradle.exps.Push(qr.Params.(string))
			cradle.ops.Push(cradle.dom)

			cradle.deps[Where] = struct{}{}
		}
		case And:
			{
				if _, ok := cradle.deps[Where]; !ok {
					return lucyErr.QueryDependencyNotSatisfied
				}
				cradle.exps.Push(qr.Params.(Exp))
				cradle.ops.Push(cradle.dom)
			}
		case AndStr:{
			if _, ok := cradle.deps[Where]; !ok {
				return lucyErr.QueryDependencyNotSatisfied
			}
			cradle.exps.Push(qr.Params.(string))
			cradle.ops.Push(cradle.dom)
		}
		case Or:
			{
				if _, ok := q.cradle.deps[Where]; !ok {
					return lucyErr.QueryDependencyNotSatisfied
				}
				cradle.exps.Push(qr.Params.(Exp))
				cradle.ops.Push(cradle.dom)
			}
		case OrStr:{
			{
				if _, ok := q.cradle.deps[Where]; !ok {
					return lucyErr.QueryDependencyNotSatisfied
				}
				cradle.exps.Push(qr.Params.(string))
				cradle.ops.Push(cradle.dom)
			}
		}
		case SetTarget:
			{
				/* If the 'Where' clause is used in conjunction with 'SetTarget (aka) Find' ,
				   then ignore params passed by query, otherwise do not ignore.
				 */

				if _, ok := q.cradle.deps[Where]; ok {
					cradle.ops.Push(SetTarget)
				} else {
					cradle.ops.Push(Where)
					cradle.exps.Push(qr.Params.(Exp))
					cradle.ops.Push(SetTarget)
				}

				cradle.out = qr.Output
			}
		case MiscNodeName: {
			cradle.ops.Push(MiscNodeName)
			cradle.exps.Push(qr.Params)
		}
		}

		cradle.pdom = cradle.dom
	}

	return nil
}
