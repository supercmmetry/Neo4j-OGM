package lucy

import "reflect"

type Layer interface {
	AttachTo(l *Database)
	StartTransaction()
	Sync() error
	AddRuntime(rt QueryRuntime)
	GetRuntime() QueryRuntime
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

func (l *Database) Find(param interface{}) *Database {
	if l.Error != nil {
		return l
	}
	l.addQuery(Query{FamilyType: SetTarget, Params: Marshal(param), Output: param})
	l.Error = l.layer.Sync()

	return l
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

func (l *Database) Create(params interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: Creation, Params: Marshal(params), Output: params})
	l.Error = l.layer.Sync()

	return l
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

func (l *Database) Set(I_ interface{}, I ...interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(I_).Kind() == reflect.Ptr && reflect.TypeOf(I_).Elem().Kind() == reflect.Struct {
		l.addQuery(Query{FamilyType: Updation, Params: Marshal(I_), Output: I_})
	} else if reflect.TypeOf(I_).Kind() == reflect.Struct {
		l.addQuery(Query{FamilyType: Updation, Params: Marshal(I_), Output: I_})
	} else if reflect.TypeOf(I_).Kind() == reflect.String {
		l.addQuery(Query{FamilyType: UpdationStr, Params: SFormat(I_.(string), I)})
	} else {
		l.Error = Error(UnrecognizedExpression)
	}

	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Model(i interface{}) *Database {
	if l.Error != nil {
		return l
	}

	if reflect.TypeOf(i).Kind() == reflect.Struct {
		l.addQuery(Query{FamilyType: Model, Params: reflect.TypeOf(i).Name()})
	} else if reflect.TypeOf(i).Kind() == reflect.Ptr && reflect.TypeOf(i).Elem().Kind() == reflect.Struct {
		l.addQuery(Query{FamilyType: Model, Params: reflect.TypeOf(i).Elem().Name()})
	} else {
		l.Error = Error(UnrecognizedExpression)
	}

	return l
}

func (l *Database) Delete() *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: Deletion})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Relate(I interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: RelationX, Params: I})
	return l
}

func (l *Database) To(I interface{}) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: RelationY, Params: I})
	return l
}

func (l *Database) By(relName string) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType: By, Params: relName})
	l.Error = l.layer.Sync()

	return l
}

func (l *Database) Relation(relName string) *Database {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{FamilyType:MTRelation, Params: Exp{"relation": relName}})

	return l
}

func (l *Database) Close() *Database {
	l.Error = l.layer.GetRuntime().Close()
	return l
}
