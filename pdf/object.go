package pdf

type Object interface {
	String() string
}

type NullObject struct {}

func NewNullObject() *NullObject {
	return &NullObject{}
}

func (obj *NullObject) String() string {
	return "null"
}

type IndirectObject struct {
	Number int64
	Generation int64
	Value Object
	Stream []byte
}

func NewIndirectObject(number int64) *IndirectObject {
	return &IndirectObject{number, 0, NewNullObject(), nil}
}
