package lucy

import (
	"fmt"
)

/* The QueryCradle is responsible for storing expressions and operators
parsed from the OGM chain. The QueryCradle is then directly passed to the
dialect-specific runtime, to generate queries.
 */

type QueryCradle struct {
	Exps               Queue
	Ops                Queue
	family, prevFamily FamilyType
	deps               map[FamilyType]struct{}
	Out                interface{}
}

func (c *QueryCradle) init() {
	c.family = Unknown
	c.prevFamily = Unknown
	c.Exps.Init()
	c.Ops.Init()
	c.deps = make(map[FamilyType]struct{})
}

/*
The QueryRuntime is entirely dialect-specific and is used to translate QueryCradle in order to generate
dialect-specific queries. It is also responsible for executing the generated queries.
 */

type QueryRuntime interface {
	CheckForInjection(expStr string) (uint, bool)
	Compile(cradle *QueryCradle) (string, error)
	Execute(query string, target interface{}) error
}

/*
The QueryEngine is responsible fo parsing the OGM chain to generate the QueryCradle.
OGM-chain specific functions are carried out here.
 */

type QueryEngine struct {
	queue             *Queue
	hasStarted        bool
	isTransaction     bool
	checkForInjection bool
	cradle            *QueryCradle
	Runtime           QueryRuntime
}

func (q *QueryEngine) ToggleInjectionCheck() {
	q.checkForInjection = !q.checkForInjection
}

func (q *QueryEngine) AddRuntime(rt QueryRuntime) {
	q.Runtime = rt
}

func (q *QueryEngine) NewQueryEngine() Layer {
	q.checkForInjection = true
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

		cradle.family = qr.FamilyType
		switch qr.FamilyType {
		case Where:
			{
				if cradle.prevFamily == Where {
					return Error(CorruptedQueryChain)
				}
				param := qr.Params.(string)

				if q.checkForInjection {
					if s, ok := q.Runtime.CheckForInjection(param); ok {
						return Error(QueryInjection, Severity(s))
					}
				}

				cradle.Exps.Push(param)
				cradle.Ops.Push(cradle.family)

				cradle.deps[Where] = struct{}{}
			}
		case And:
			{
				if _, ok := cradle.deps[Where]; !ok {
					return Error(UnsatisfiedDependency)
				}
				param := qr.Params.(string)

				if q.checkForInjection {
					if s, ok := q.Runtime.CheckForInjection(param); ok {
						return Error(QueryInjection, Severity(s))
					}
				}

				cradle.Exps.Push(param)
				cradle.Ops.Push(cradle.family)
			}
		case Or:
			{
				{
					if _, ok := q.cradle.deps[Where]; !ok {
						return Error(UnsatisfiedDependency)
					}

					param := qr.Params.(string)

					if q.checkForInjection {
						if s, ok := q.Runtime.CheckForInjection(param); ok {
							return Error(QueryInjection, Severity(s))
						}
					}

					cradle.Exps.Push(param)
					cradle.Ops.Push(cradle.family)
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
					exp := qr.Params.(Exp)
					for k, v := range exp {
						exp[k] = Format("?", v)
					}
					cradle.Exps.Push(exp)
					cradle.Ops.Push(SetTarget)
				}

				cradle.Out = qr.Output
			}
		case MiscNodeName:
			{
				cradle.Ops.Push(MiscNodeName)
				cradle.Exps.Push(qr.Params)
				cradle.deps[MiscNodeName] = struct{}{}
			}
		case Creation:
			cradle.Ops.Push(cradle.family)

			exp := qr.Params.(Exp)
			for k, v := range exp {
				exp[k] = Format("?", v)
			}
			cradle.Exps.Push(exp)
			cradle.Out = qr.Output

		}

		cradle.prevFamily = cradle.family
	}


	query, err := q.Runtime.Compile(q.cradle)
	if err != nil {
		q.cradle.init()
		return err
	}

	fmt.Println("Generated query: ", query)

	if err := q.Runtime.Execute(query, q.cradle.Out); err != nil {
		q.cradle.init()
		return err
	}

	q.cradle.init()
	return nil
}
