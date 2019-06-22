package pdf

import (
	"fmt"
)

type Action Dictionary

func (a Action) Extract(output *Output) {
	d := Dictionary(a)

	// filespecification can be in either F or Win[F]
	if f, err := d.GetString("F"); err == nil {
		fmt.Fprintf(output.Files, "unknown:%s\n", f)
	} else if f, err := d.GetDictionary("F"); err == nil {
		File(f).Extract(output)
	}
	if win, err := d.GetDictionary("Win"); err == nil {
		File(win).Extract(output)
	}
}
