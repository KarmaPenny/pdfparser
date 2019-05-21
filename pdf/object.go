package pdf

import (
	"fmt"
	"io"
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

func (obj *IndirectObject) Fprint(w io.Writer) {
	value := obj.Value.String()
	if value != "null" {
		fmt.Fprintf(w, "%d %d obj\n", obj.Number, obj.Generation)
		fmt.Fprintln(w, obj.Value.String())
		if obj.Stream != nil {
			fmt.Fprintln(w, "stream")
			fmt.Fprintln(w, string(obj.Stream))
			fmt.Fprintln(w, "endstream")
		}
		fmt.Fprintln(w, "endobj")
	}
}
