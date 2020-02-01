package lucy

type ObjectMapping interface {
	Find(param interface{}) error
}
