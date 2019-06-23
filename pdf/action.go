package pdf

import (
	"fmt"
)

type Action Dictionary

func (a Action) Extract(output *Output) {
	d := Dictionary(a)

	isCommand := false
	if s, err := d.GetName("S"); err == nil && string(s) == "Launch" {
		isCommand = true
	}

	// filespecification can be in either F or Win[F]
	if f, err := d.GetString("F"); err == nil {
		fmt.Fprintf(output.Files, "%s:%s\n", unknownHash, f)
	} else if f, err := d.GetDictionary("F"); err == nil {
		File(f).Extract(output, isCommand)
	}
	if win, err := d.GetDictionary("Win"); err == nil {
		File(win).Extract(output, isCommand)
	}
}
