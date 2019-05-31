package pdf

import (
	"fmt"
	"strings"
)

type Object interface {
	String() string
}

type IndirectObject struct {
	Number int
	Generation int
	Value Object
	Stream []byte
}

func NewIndirectObject(number int) *IndirectObject {
	return &IndirectObject{number, 0, KEYWORD_NULL, nil}
}

func (object *IndirectObject) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d %d obj\n%s\n", object.Number, object.Generation, object.Value))
	if object.Stream != nil {
		sb.WriteString(fmt.Sprintf("stream\n%s\nendstream\n", string(object.Stream)))
	}
	sb.WriteString("endobj\n")
	return sb.String()
}
