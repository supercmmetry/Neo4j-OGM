package lucy

import "reflect"

type Layer interface {
	AttachTo(l *Database)
	StartTransaction()
	Sync() error
	AddRuntime(rt QueryRuntime)
	ToggleInjectionCheck()
}
type Database struct {
	Queue Queue
	Error error
	layer Layer
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

	if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: Where, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = Error(UnrecognizedExpression)
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

	if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: And, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = Error(UnrecognizedExpression)
	}

	return l
}

func (l *Database) Or(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: Or, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = Error(UnrecognizedExpression)
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

