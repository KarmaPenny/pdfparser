package pdf

import (
	"bytes"
	"strings"
)

var FontDefault *Font = &Font{map[int]string{}, 1}

type Font struct {
	Cmap map[int]string
	Width int
}

func NewFont(d Dictionary) *Font {
	cmap_string, _ := d.GetStream("ToUnicode")
	cmap := []byte(cmap_string)

	// create new font object
	font := &Font{map[int]string{}, 1}

	// create parser for parsing cmap
	parser := NewParser(bytes.NewReader(cmap), nil)

	for {
		// read next command
		command, operands, err := parser.ReadCommand()
		if err == ReadError {
			break
		}

		if command == KEYWORD_BEGIN_BF_RANGE {
			count, _ := operands.GetInt(len(operands) - 1)
			for i := 0; i < count; i++ {
				start_b := parser.ReadHexString(noDecryptor)
				if start_b == "" {
					break
				}
				font.Width = len([]byte(start_b))
				start := BytesToInt([]byte(start_b))

				end_b := parser.ReadHexString(noDecryptor)
				if end_b == "" {
					break
				}
				end := BytesToInt([]byte(end_b))

				value := parser.ReadHexString(noDecryptor)
				if value == "" {
					break
				}

				for j := start; j <= end; j++ {
					font.Cmap[j] = string(value)
				}
			}
		} else if command == KEYWORD_BEGIN_BF_CHAR {
			count, _ := operands.GetInt(len(operands) - 1)
			for i := 0; i < count; i++ {
				key_b := parser.ReadHexString(noDecryptor)
				if key_b == "" {
					break
				}
				font.Width = len([]byte(key_b))
				key := BytesToInt([]byte(key_b))

				value := parser.ReadHexString(noDecryptor)
				if value == "" {
					break
				}

				font.Cmap[key] = string(value)
			}
		}
	}

	return font
}

func (font *Font) Decode(b []byte) string {
	var s strings.Builder
	for i := 0; i + font.Width <= len(b); i += font.Width {
		bs := b[i:i + font.Width]
		k := BytesToInt(bs)
		if v, ok := font.Cmap[k]; ok {
			s.WriteString(v)
		} else {
			s.WriteString(string(bs))
		}
	}
	return s.String()
}
