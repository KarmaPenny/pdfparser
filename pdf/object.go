package pdf

import (
	"fmt"
	"strings"
)

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

func (obj *IndirectObject) String() string {
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%d %d obj\n", obj.Number, obj.Generation))
	b.WriteString(obj.Value.String())
	if obj.Stream != nil {
		b.WriteString("\nstream\n")
		b.WriteString(string(obj.Stream))
		b.WriteString("\nendstream")
	}
	b.WriteString("\nendobj\n")
	return b.String()
}
