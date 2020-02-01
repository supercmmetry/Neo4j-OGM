package lucy

import (
	lucyErr "lucy/errors"
	"reflect"
)

func Marshal(v interface{}) map[string]interface{} {
	vtype := reflect.TypeOf(v)
	vvalue := reflect.ValueOf(v)

	if vtype.Kind() != reflect.Struct {
		vtype = reflect.TypeOf(v).Elem()
		vvalue = reflect.ValueOf(v).Elem()
	}

	tagMap := make(map[string]interface{})

	for i := 0; i < vtype.NumField(); i++ {
		if tagName, ok := vtype.Field(i).Tag.Lookup("lucy"); ok {
			tagMap[tagName] = vvalue.Field(i).Interface()
		}
	}

	return tagMap
}

func IsEndDomain(d DomainType) bool {
	return d == SetTarget || d == Creation || d == Deletion || d == Updation
}

type Queue struct {
	elements *[]interface{}
}

func (q *Queue) Init() {
	elements := make([]interface{}, 0)
	q.elements = &elements
}

func (q *Queue) Push(elem interface{}) {
	*q.elements = append(*q.elements, elem)
}

func (q *Queue) Get() (interface{}, error) {
	if (len(*q.elements)) == 0 {
		return Unknown, lucyErr.EmptyQueue
	}
	elem := (*q.elements)[0]
	*q.elements = (*q.elements)[1:]
	return elem, nil
}

func (q *Queue) IsEmpty() bool {
	return len(*q.elements) == 0
}
