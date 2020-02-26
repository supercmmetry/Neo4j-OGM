package lucy

import (
	"github.com/supercmmetry/lucy/types"
	"reflect"
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
	AllowEmptyResult   bool
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
	Execute(query string, cradle *QueryCradle, target interface{}) error
	Close() error
	Commit() error
	Rollback() error
	BeginTransaction() error
	CloseTransaction() error
	Clone() QueryRuntime
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

func (q *QueryEngine) GetRuntime() QueryRuntime {
	return q.Runtime
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
		cradle.AllowEmptyResult = true

		qri, err := q.queue.Get()
		if err != nil {
			return err
		}

		qr := qri.(Query)

		cradle.family = qr.FamilyType
		switch qr.FamilyType {
		case Model:
			cradle.Exps.Push(qr.Params.(string))
			cradle.Ops.Push(cradle.family)
			cradle.deps[Model] = struct{}{}
			break
		case Where:
			if _, ok := q.cradle.deps[Where]; ok {
				return Error(CorruptedQueryChain, "Where() used more than once in query chain")
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

			// Destroy multi-target chain if where clause is used.
			cradle.Out = nil
			break
		case And:
			if _, ok := cradle.deps[Where]; !ok {
				return Error(UnsatisfiedDependency, "missing: Where()")
			}
			param := qr.Params.(string)

			if q.checkForInjection {
				if s, ok := q.Runtime.CheckForInjection(param); ok {
					return Error(QueryInjection, Severity(s))
				}
			}

			cradle.Exps.Push(param)
			cradle.Ops.Push(cradle.family)
			break
		case Or:
			if _, ok := q.cradle.deps[Where]; !ok {
				return Error(UnsatisfiedDependency, "missing: Where()")
			}

			param := qr.Params.(string)

			if q.checkForInjection {
				if s, ok := q.Runtime.CheckForInjection(param); ok {
					return Error(QueryInjection, Severity(s))
				}
			}

			cradle.Exps.Push(param)
			cradle.Ops.Push(cradle.family)

			break
		case SetTarget:

			cradle.Ops.Push(SetTarget)
			cradle.Out = qr.Output
			cradle.deps[SetTarget] = struct{}{}
			cradle.AllowEmptyResult = false

			break
		case Creation:
			cradle.Ops.Push(cradle.family)

			exp := qr.Params.(types.Exp)
			for k, v := range exp {
				exp[k] = Format("?", v)
			}
			cradle.Exps.Push(exp)
			cradle.Out = qr.Output
			break
		case Updation:
			cradle.AllowEmptyResult = false

			if q.cradle.Out == nil || (reflect.TypeOf(q.cradle.Out).Kind() == reflect.Ptr &&
				reflect.TypeOf(q.cradle.Out).Elem().Kind() == reflect.Slice) ||
				reflect.TypeOf(q.cradle.Out).Kind() == reflect.Slice {

				if _, ok := q.cradle.deps[Where]; !ok {
					return Error(UnsatisfiedDependency, "missing: Where()")
				}
			}

			cradle.Ops.Push(cradle.family)

			exp := qr.Params.(types.Exp)
			for k, v := range exp {
				exp[k] = Format("?", v)
			}
			cradle.Exps.Push(exp)
			break
		case UpdationStr:
			cradle.AllowEmptyResult = false

			if q.cradle.Out == nil || (reflect.TypeOf(q.cradle.Out).Kind() == reflect.Ptr &&
				reflect.TypeOf(q.cradle.Out).Elem().Kind() == reflect.Slice) ||
				reflect.TypeOf(q.cradle.Out).Kind() == reflect.Slice {

				if _, ok := q.cradle.deps[Where]; !ok {
					return Error(UnsatisfiedDependency, "missing: Where()")
				}
				if _, ok := q.cradle.deps[Model]; !ok {
					return Error(UnsatisfiedDependency, "missing: Model()")
				}
			}

			param := qr.Params.(string)

			if q.checkForInjection {
				if s, ok := q.Runtime.CheckForInjection(param); ok {
					return Error(QueryInjection, Severity(s))
				}
			}

			cradle.Exps.Push(param)
			cradle.Ops.Push(cradle.family)
			break
		case Deletion:
			if q.cradle.Out == nil || (reflect.TypeOf(q.cradle.Out).Kind() == reflect.Ptr &&
				reflect.TypeOf(q.cradle.Out).Elem().Kind() == reflect.Slice) ||
				reflect.TypeOf(q.cradle.Out).Kind() == reflect.Slice {

				if _, ok := q.cradle.deps[Where]; !ok {
					return Error(UnsatisfiedDependency, "missing: Where()")
				}
				if _, ok := q.cradle.deps[Model]; !ok {
					return Error(UnsatisfiedDependency, "missing: Model()")
				}
			}

			cradle.Ops.Push(cradle.family)
			break
		case RelationX:
			// nuke existing multi-target.
			cradle.Out = nil

			cradle.Ops.Push(cradle.family)
			cradle.Exps.Push(qr.Params)
			cradle.deps[RelationX] = struct{}{}
			break
		case RelationY:
			if _, ok := cradle.deps[RelationX]; !ok {
				return Error(UnsatisfiedDependency, "missing: Relate()")
			}

			cradle.Ops.Push(cradle.family)
			cradle.Exps.Push(qr.Params)
			cradle.deps[RelationY] = struct{}{}
			break
		case By:
			if _, ok := cradle.deps[RelationY]; !ok {
				return Error(UnsatisfiedDependency, "missing: To()")
			}

			cradle.Ops.Push(cradle.family)
			cradle.Exps.Push(qr.Params)
			break

		case MTRelation:

			// Use Relation() only after Find()
			if q.cradle.Out == nil {
				return Error(UnsatisfiedDependency, "missing: Find()")
			}

			cradle.Ops.Push(cradle.family)

			exp := qr.Params.(types.Exp)
			exp["out"] = cradle.Out
			cradle.Exps.Push(exp)

			break
		}

		cradle.prevFamily = cradle.family

	}

	query, err := q.Runtime.Compile(q.cradle)
	if err != nil {
		q.cradle.init()
		return err
	}

	// fmt.Println("Generated query: ", query)

	if err := q.Runtime.Execute(query, q.cradle, q.cradle.Out); err != nil {
		q.cradle.init()
		return err
	}

	if _, ok := cradle.deps[SetTarget]; ok {
		if cradle.Out == nil {
			return Error(NoRecordsFound)
		}
	}

	q.cradle.init()
	return nil
}
