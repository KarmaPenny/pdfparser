package pdf

import (
	"bytes"
	"io"
	"strings"
)

type Page Dictionary

func (page Page) ExtractText(w io.Writer) {
	d := Dictionary(page)

	// load fonts
	font_map := map[string]*Font{}
	resources, _ := d.GetDictionary("Resources")
	fonts, _ := resources.GetDictionary("Font")
	for font := range fonts {
		font_info, _ := fonts.GetDictionary(font)
		font_map[font] = NewFont(font_info)
	}

	// get contents
	contents, _ := d.GetStream("Contents")

	// create parser for parsing contents
	page_parser := NewParser(bytes.NewReader(contents))

	for {
		// read next command
		command, _, err := page_parser.ReadCommand()
		if err == ErrorRead {
			break
		}

		// start of text block
		if command == KEYWORD_TEXT {
			// initial font is none
			current_font := FontDefault

			for {
				command, operands, err := page_parser.ReadCommand()
				// stop if end of stream or end of text block
				if err == ErrorRead || command == KEYWORD_TEXT_END {
					break
				}

				// handle font changes
				if command == KEYWORD_TEXT_FONT {
					font_name, _ := operands.GetName(len(operands) - 2)
					if font, ok := font_map[font_name]; ok {
						current_font = font
					} else {
						current_font = FontDefault
					}
				} else if command == KEYWORD_TEXT_SHOW_1 || command == KEYWORD_TEXT_SHOW_2 || command == KEYWORD_TEXT_SHOW_3 {
					// decode text with current font font
					s, _ := operands.GetString(len(operands) - 1)
					io.WriteString(w, current_font.Decode([]byte(s)))
					io.WriteString(w, "\n")
				} else if command == KEYWORD_TEXT_POSITION {
					// decode positioned text with current font
					var sb strings.Builder
					a, _ := operands.GetArray(len(operands) - 1)
					for i := 0; i < len(a); i += 2 {
						s, _ := a.GetString(i)
						sb.WriteString(string(s))
					}
					io.WriteString(w, current_font.Decode([]byte(sb.String())))
					io.WriteString(w, "\n")
				}
			}
		}
	}
}
