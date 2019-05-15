package pdf

import (
	"bytes"
	"fmt"
	"strings"
)

type XrefEntry struct {
	Offset int64
	Generation int64
	Type int64
}

func NewXrefEntry(offset int64, generation int64, type_value int64) *XrefEntry {
	return &XrefEntry{offset, generation, type_value}
}

func (entry *XrefEntry) String() string {
	return fmt.Sprintf("%d - %d - %d", entry.Type, entry.Offset, entry.Generation)
}

type Token struct {
	*bytes.Buffer
	IsNumber bool
}

func NewToken(b byte) *Token {
	return &Token{bytes.NewBuffer([]byte{b}), false}
}

func NewTokenString(s string) *Token {
	return &Token{bytes.NewBufferString(s), false}
}

func (token Token) String() string {
	return string(token.Bytes())
}

type Array []fmt.Stringer

func (a Array) String() string {
	var s strings.Builder
	s.WriteString("[")
	for _, value := range a {
		s.WriteString(value.String())
		s.WriteString(" ")
	}
	s.WriteString("]")
	return s.String()
}

type Dictionary map[string]fmt.Stringer

func (d Dictionary) String() string {
	var s strings.Builder
	s.WriteString("<<")
	for key, value := range d {
		s.WriteString(key)
		s.WriteString(" ")
		s.WriteString(value.String())
	}
	s.WriteString(">>")
	return s.String()
}

type Reference struct {
	Number int64
	Generation int64
}

func NewReference(number int64, generation int64) *Reference {
	return &Reference{number, generation}
}

func (reference *Reference) String() string {
	return fmt.Sprintf("%d %d R", reference.Number, reference.Generation)
}
