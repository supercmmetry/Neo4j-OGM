package lucy

import (
	e "lucy/errors"
)


type QueryCradle struct {
	Exps      Queue
	Ops       Queue
	dom, pdom DomainType
	deps      map[DomainType]struct{}
	Out       interface{}
}

func (c *QueryCradle) init() {
	c.dom = Unknown
	c.pdom = Unknown
	c.Exps.Init()
	c.Ops.Init()
	c.deps = make(map[DomainType]struct{})
}

type QueryRuntime interface {
	CheckForInjection(expStr string) (uint, bool)
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
					return e.Error(e.CorruptedQueryChain)
				}
				exp := qr.Params.(Exp)
				for k,v := range exp {
					exp[k] = Format("?", v) // Sanitize values

					// Detect injection in keys
					if s, ok := q.Runtime.CheckForInjection(k); ok {
						return e.Error(e.QueryInjection, e.Severity(s))
					}
				}
				cradle.Exps.Push(exp)
				cradle.Ops.Push(Where)

				cradle.deps[Where] = struct{}{}
			}
		case WhereStr: {
			if cradle.pdom == Where {
				return e.Error(e.CorruptedQueryChain)
			}
			param := qr.Params.(string)

			if s, ok := q.Runtime.CheckForInjection(param); ok {
				return e.Error(e.QueryInjection, e.Severity(s))
			}

			cradle.Exps.Push(param)
			cradle.Ops.Push(cradle.dom)

			cradle.deps[Where] = struct{}{}
		}
		case And:
			{
				if _, ok := cradle.deps[Where]; !ok {
					return e.Error(e.UnsatisfiedDependency)
				}
				exp := qr.Params.(Exp)
				for k,v := range exp {
					exp[k] = Format("?", v) // Sanitize values

					// Detect injection in keys
					if s, ok := q.Runtime.CheckForInjection(k); ok {
						return e.Error(e.QueryInjection, e.Severity(s))
					}
				}
				cradle.Exps.Push(exp)
				cradle.Ops.Push(cradle.dom)
			}
		case AndStr:{
			if _, ok := cradle.deps[Where]; !ok {
				return e.Error(e.UnsatisfiedDependency)
			}
			param := qr.Params.(string)

			if s, ok := q.Runtime.CheckForInjection(param); ok {
				return e.Error(e.QueryInjection, e.Severity(s))
			}

			cradle.Exps.Push(param)
			cradle.Ops.Push(cradle.dom)
		}
		case Or:
			{
				if _, ok := q.cradle.deps[Where]; !ok {
					return e.Error(e.UnsatisfiedDependency)
				}
				exp := qr.Params.(Exp)
				for k,v := range exp {
					exp[k] = Format("?", v) // Sanitize values

					// Detect injection in keys
					if s, ok := q.Runtime.CheckForInjection(k); ok {
						return e.Error(e.QueryInjection, e.Severity(s))
					}
				}
				cradle.Exps.Push(exp)
				cradle.Ops.Push(cradle.dom)
			}
		case OrStr:{
			{
				if _, ok := q.cradle.deps[Where]; !ok {
					return e.Error(e.UnsatisfiedDependency)
				}

				param := qr.Params.(string)

				if s, ok := q.Runtime.CheckForInjection(param); ok {
					return e.Error(e.QueryInjection, e.Severity(s))
				}

				cradle.Exps.Push(param)
				cradle.Ops.Push(cradle.dom)
			}
		}
		case SetTarget:
			{
				/* If the 'Where' clause is used in conjunction with 'SetTarget (aka) Find' ,
				   then ignore params passed by query, otherwise do not ignore.
				*/

				if _, ok := q.cradle.deps[Where]; ok {
					cradle.Ops.Push(SetTarget)
				} else {
					cradle.Ops.Push(Where)
					cradle.Exps.Push(qr.Params.(Exp))
					cradle.Ops.Push(SetTarget)
				}

				cradle.Out = qr.Output
			}
		case MiscNodeName: {
			cradle.Ops.Push(MiscNodeName)
			cradle.Exps.Push(qr.Params)
		}
		}

		cradle.pdom = cradle.dom
	}

	return nil
}