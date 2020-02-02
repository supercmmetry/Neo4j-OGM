package lucy

import (
	"fmt"
	lucyErr "lucy/errors"
	"reflect"
	"strconv"
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

func Format(format string, I ...interface{}) string {
	newStr := ""
	index := 0
	for _, chr := range format {
		if chr == '?' {
			index += 1
			i := index - 1
			switch reflect.TypeOf(I[i]).Kind() {
			case reflect.String:
				{
					subStr := ""
					targStr := I[i].(string)
					for _, c := range targStr {
						if c == '\'' || c == '"' {
							subStr += "\\"
						}
						subStr += string(c)
					}
					newStr += "'" + subStr + "'"
				}
			case reflect.Int: {
				newStr += strconv.Itoa(I[i].(int))
			}
			case reflect.Int64: {
				newStr += strconv.FormatInt(I[i].(int64), 10)
			}
			case reflect.Int32: {
				newStr += strconv.FormatInt(int64(I[i].(int32)), 10)
			}
			case reflect.Int16: {
				newStr += strconv.FormatInt(int64(I[i].(int16)), 10)
			}
			case reflect.Int8: {
				newStr += strconv.FormatInt(int64(I[i].(int8)), 10)
			}
			case reflect.Uint: {
				newStr += strconv.FormatUint(uint64(I[i].(uint)), 10)
			}
			case reflect.Uint8: {
				newStr += strconv.FormatUint(uint64(I[i].(uint8)), 10)
			}
			case reflect.Uint16: {
				newStr += strconv.FormatUint(uint64(I[i].(uint16)), 10)
			}
			case reflect.Uint32: {
				newStr += strconv.FormatUint(uint64(I[i].(uint32)), 10)
			}
			case reflect.Uint64: {
				newStr += strconv.FormatUint(I[i].(uint64), 10)
			}
			case reflect.Float32: {
				newStr += fmt.Sprintf("%f", I[i].(float32))
			}
			case reflect.Float64: {
				newStr += fmt.Sprintf("%f", I[i].(float64))
			}
			}
		} else {
			newStr += string(chr)
		}
	}
	return newStr
}

type Queue struct {
	elements *[]interface{}
}

func (q *Queue) Init() {
	elements := make([]interface{}, 0)
	q.elements = &elements
}

func (q *Queue) GetAll() *[]interface{} {
	return q.elements
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
