package lucy

import (
	e "github.com/supercmmetry/lucy/errors"
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
	ToggleInjectionCheck()
}

type KeyValuePair struct {
	Key   string
	Value interface{}
}

type Database struct {
	Queue         Queue
	Error         error
	layer         Layer
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

func (l *Database) ToggleInjectionCheck() {
	l.layer.ToggleInjectionCheck()
}

func (l *Database) Find(param interface{}) error {
	if l.Error != nil {
		return l.Error
	}
	l.addQuery(Query{FamilyType: SetTarget, Params: Marshal(param), Output: param})
	l.Error = l.layer.Sync()

	return l.Error
}

func (l *Database) Where(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_) == reflect.TypeOf(Exp{}) {
		Exp := I_
		l.addQuery(Query{FamilyType: Where, Params: Exp})
	} else if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: WhereStr, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = e.Error(e.UnrecognizedExpression)
	}

	return l
}

func (l *Database) Create(params interface{}) error {
	if l.Error != nil {
		return l.Error
	}

	l.addQuery(Query{FamilyType: Creation, Params: Marshal(params), Output: params})
	l.Error = l.layer.Sync()

	return l.Error
}

func (l *Database) And(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_) == reflect.TypeOf(Exp{}) {
		Exp := I_
		l.addQuery(Query{FamilyType: And, Params: Exp})
	} else if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: AndStr, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = e.Error(e.UnrecognizedExpression)
	}

	return l
}

func (l *Database) Or(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_) == reflect.TypeOf(Exp{}) {
		Exp := I_
		l.addQuery(Query{FamilyType: Or, Params: Exp})
	} else if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: OrStr, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = e.Error(e.UnrecognizedExpression)
	}

	return l
}

func (l *Database) By(name string) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: MiscNodeName, Params: name})

	return l
}

func (l *Database) Model(i interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: Model, Params: Marshal(i)})

	return l
}

func (l *Database) Begin() *Database {
	l.layer.StartTransaction()
	return l
}
