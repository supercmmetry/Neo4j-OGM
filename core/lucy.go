package lucy

type Lucy struct {
	Engine  Layer
	db      *Database
	runtime QueryRuntime
}

func (l *Lucy) DB() *Database {
	l.Engine = (&QueryEngine{}).NewQueryEngine()
	l.db = &Database{}
	l.Engine.AttachTo(l.db)
	l.db.AddRuntime(l.runtime)
	return l.db
}

func (l *Lucy) AddRuntime(rt QueryRuntime) {
	l.runtime = rt
}

type Expr map[string]interface{}

type Layer interface {
	AttachTo(l *Database)
	Start()
	StartTransaction()
	Stop()
	Sync() error
	AddRuntime(rt QueryRuntime)
}

type KeyValuePair struct {
	Key   string
	Value interface{}
}

type Database struct {
	Queue         QueryQueue
	Error         error
	layer         Layer
	isTransaction bool
}

func (l *Database) addQuery(query Query) {
	l.Queue.Push(query)
}

func (l *Database) AddRuntime(rt QueryRuntime) {
	l.layer.AddRuntime(rt)
}

func (l *Database) SetLayer(layer Layer) {
	l.layer = layer
	l.Queue.Init()
}

func (l *Database) Find(param interface{}) *Database {
	l.Where(Marshal(param))

	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: SetTarget, Output: param})
	l.layer.Start()
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Where(expr Expr) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Where, Params: expr})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Create(params interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Creation, Params: Marshal(params)})
	l.layer.Start()
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) And(expr Expr) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: And, Params: expr})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Or(expr Expr) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Or, Params: expr})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Begin() *Database {
	l.layer.StartTransaction()
	return l
}
