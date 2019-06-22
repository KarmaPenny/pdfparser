package pdf

import (
	"fmt"
	"io"
)

type Action Dictionary

func (a Action) Extract(file_output, url_output io.Writer, extract_dir string) {
	d := Dictionary(a)

	// filespecification can be in either F or Win[F]
	if f, err := d.GetString("F"); err == nil {
		fmt.Fprintf(file_output, "unknown:%s\n", f)
	} else if f, err := d.GetDictionary("F"); err == nil {
		File(f).Extract(file_output, url_output, extract_dir)
	}
	if win, err := d.GetDictionary("Win"); err == nil {
		if f, err := win.GetString("F"); err == nil {
			fmt.Fprintf(file_output, "unknown:%s\n", f)
		} else if f, err := win.GetDictionary("F"); err == nil {
			File(f).Extract(file_output, url_output, extract_dir)
		}
	}
}
