package lucy

import (
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
