package pdf

import (
	"bytes"
)

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
