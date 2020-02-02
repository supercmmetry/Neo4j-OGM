package lucy

import (
	lucyErr "lucy/errors"
	"reflect"
)

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

type Exp map[string]interface{}

type Layer interface {
	AttachTo(l *Database)
	StartTransaction()
	Sync() error
	AddRuntime(rt QueryRuntime)
}

type KeyValuePair struct {
	Key   string
	Value interface{}
}

type Database struct {
	Queue         Queue
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
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: SetTarget, Params: Marshal(param), Output: param})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Where(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_) == reflect.TypeOf(Exp{}) {
		Exp := I_
		l.addQuery(Query{DomainType: Where, Params: Exp})
	} else if reflect.TypeOf(I_) == reflect.TypeOf("") {
		l.addQuery(Query{DomainType: WhereStr, Params: Format(I_.(string), I)})
	} else {
		l.Error = lucyErr.ExpressionNotRecognized
	}

	return l
}

func (l *Database) Create(params interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Creation, Params: Marshal(params)})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) And(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_) == reflect.TypeOf(Exp{}) {
		Exp := I_
		l.addQuery(Query{DomainType: And, Params: Exp})
	} else if reflect.TypeOf(I_) == reflect.TypeOf("") {
		l.addQuery(Query{DomainType: AndStr, Params: Format(I_.(string), I)})
	} else {
		l.Error = lucyErr.ExpressionNotRecognized
	}

	return l
}

func (l *Database) Or(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_) == reflect.TypeOf(Exp{}) {
		Exp := I_
		l.addQuery(Query{DomainType: Or, Params: Exp})
	} else if reflect.TypeOf(I_) == reflect.TypeOf("") {
		l.addQuery(Query{DomainType: OrStr, Params: Format(I_.(string), I)})
	} else {
		l.Error = lucyErr.ExpressionNotRecognized
	}

	return l
}

func (l *Database) By(name string) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: MiscNodeName, Params: name})

	return l
}

func (l *Database) Model(i interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Model, Params: Marshal(i)})

	return l
}

func (l *Database) Begin() *Database {
	l.layer.StartTransaction()
	return l
}
