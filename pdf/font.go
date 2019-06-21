package pdf

import (
	//"bytes"
)

type Font struct {
	Cmap map[string]string
	Width int
}

func NewFont(cmap []byte) *Font {
	font := &Font{map[string]string{}, 1}

	//parser := NewPdf(bytes.NewReader(cmap))

	return font
}
