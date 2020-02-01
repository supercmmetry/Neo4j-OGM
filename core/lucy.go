package lucy

type Expr map[string]interface{}

type Layer interface {
	AttachTo(l *Lucy)
	Start()
	Stop()
	Sync() error
}

type KeyValuePair struct {
	Key   string
	Value interface{}
}

type Lucy struct {
	Queue   QueryQueue
	Mapping ObjectMapping
	Error   error
	layer   Layer
}

func (l *Lucy) addQuery(query Query) {
	l.Queue.Push(query)
}

func (l *Lucy) SetLayer(layer Layer) {
	l.layer = layer
	l.Queue.Init()
}

func (l *Lucy) Find(param interface{}) *Lucy {
	l.Where(Marshal(param))

	if l.Error != nil {
		return l
	}


	l.addQuery(Query{DomainType: SetTarget, Output: param})
	l.layer.Start()
	l.Error = l.layer.Sync()

	return l
}


func (l *Lucy) Where(expr Expr) *Lucy {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Where, Params: expr})
	l.Error = l.layer.Sync()

	return l
}

func (l *Lucy) Create(params interface{}) *Lucy {
	if l.Error != nil {
		return l
	}

	l.addQuery(Query{DomainType: Creation, Params: Marshal(params)})
	l.layer.Start()
	l.Error = l.layer.Sync()

	return l
}
