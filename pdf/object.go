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

func (object *IndirectObject) Extract(output *Output) {
	// get object dictionary
	if d, ok := object.Value.(Dictionary); ok {
		extract(d, output)
	}
}

func extract(o Object, output *Output) {
	if d, ok := o.(Dictionary); ok {
		// dump actions
		if a, ok := d.GetDictionary("A"); ok {
			Action(a).Extract(output)
		}

		// dump open action
		if open_action, ok := d.GetDictionary("OpenAction"); ok {
			Action(open_action).Extract(output)
		}

		// dump additional actions
		if aa, ok := d.GetDictionary("AA"); ok {
			for key := range aa {
				if a, ok := aa.GetDictionary(key); ok {
					Action(a).Extract(output)
				}
			}
		}

		// dump forms
		if xfa, ok := d.GetStream("XFA"); ok {
			output.DumpFile("form.xml", xfa)
		} else if xfa, ok := d.GetArray("XFA"); ok {
			var form_data strings.Builder
			for i := range xfa {
				if s, ok := xfa.GetStream(i); ok {
					form_data.WriteString(string(s))
				}
			}
			output.DumpFile("form.xml", []byte(form_data.String()))
		}

		// dump Embedded Files
		embedded_files := d.GetNameTreeMap("EmbeddedFiles")
		for i := 1; i < len(embedded_files); i += 2 {
			if f, ok := embedded_files.GetString(i); ok {
				fmt.Fprintf(output.Files, "%s:%s\n", unknownHash, f)
			} else if f, ok := embedded_files.GetDictionary(i); ok {
				File(f).Extract(output, false)
			}
		}

		// dump javascript
		if js, ok := d.GetString("JS"); ok {
			fmt.Fprintln(output.Javascript, js)
		} else if js, ok := d.GetStream("JS"); ok {
			fmt.Fprintln(output.Javascript, string(js))
		}

		// dump page text
		if pages, ok := d.GetPageTree("Pages"); ok {
			for i := range pages {
				Page(pages[i]).Extract(output)
			}
		}

		// dump URIs
		if url, ok := d.GetString("URI"); ok {
			fmt.Fprintln(output.URLs, string(url))
		} else if url, ok := d.GetDictionary("URI"); ok {
			if base, ok := url.GetString("Base"); ok {
				fmt.Fprintln(output.URLs, string(base))
			}
		}

		// dump URLs
		urls := d.GetNameTreeMap("URLS")
		for i := 0; i < len(urls); i += 2 {
			if url, ok := urls.GetString(i); ok {
				fmt.Fprintln(output.URLs, string(url))
			}
		}

		for key := range d {
			extract(d[key], output)
		}
	} else if a, ok := o.(Array); ok {
		for i := range a {
			extract(a[i], output)
		}
	}
}
