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
		if a, err := d.GetDictionary("A"); err == nil {
			Action(a).Extract(output)
		}

		// dump open action
		if open_action, err := d.GetDictionary("OpenAction"); err == nil {
			Action(open_action).Extract(output)
		}

		// dump additional actions
		if aa, err := d.GetDictionary("AA"); err == nil {
			for key := range aa {
				if a, err := aa.GetDictionary(key); err == nil {
					Action(a).Extract(output)
				}
			}
		}

		// dump forms
		if xfa, err := d.GetStream("XFA"); err == nil {
			output.DumpFile("form.xml", xfa)
		} else if xfa, err := d.GetArray("XFA"); err == nil {
			var form_data strings.Builder
			for i := range xfa {
				if s, err := xfa.GetStream(i); err == nil {
					form_data.WriteString(string(s))
				}
			}
			output.DumpFile("form.xml", []byte(form_data.String()))
		}

		// dump Embedded Files
		embedded_files := d.GetNameTreeMap("EmbeddedFiles")
		for i := 1; i < len(embedded_files); i += 2 {
			if f, err := embedded_files.GetString(i); err == nil {
				fmt.Fprintf(output.Files, "unknown:%s\n", f)
			} else if f, err := embedded_files.GetDictionary(i); err == nil {
				File(f).Extract(output, false)
			}
		}

		// dump javascript
		if js, err := d.GetString("JS"); err == nil {
			fmt.Fprintln(output.Javascript, string(js))
		} else if js, err := d.GetStream("JS"); err == nil {
			fmt.Fprintln(output.Javascript, string(js))
		}

		// dump page text
		if t, err := d.GetName("Type"); err == nil && string(t) == "Page" {
			Page(d).Extract(output)
		}

		// dump URIs
		if url, err := d.GetString("URI"); err == nil {
			fmt.Fprintln(output.URLs, string(url))
		} else if url, err := d.GetDictionary("URI"); err == nil {
			if base, err := url.GetString("Base"); err == nil {
				fmt.Fprintln(output.URLs, string(base))
			}
		}

		// dump URLs
		urls := d.GetNameTreeMap("URLS")
		for i := 0; i < len(urls); i += 2 {
			if url, err := urls.GetString(i); err == nil {
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
