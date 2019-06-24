package pdf

import (
	"fmt"
)

type Action Dictionary

func (a Action) Extract(output *Output) {
	d := Dictionary(a)

	isCommand := false
	if s, ok := d.GetName("S"); ok && s == "Launch" {
		isCommand = true
	}

	// filespecification can be in either F or Win[F]
	if f, ok := d.GetString("F"); ok {
		fmt.Fprintf(output.Files, "%s:%s\n", unknownHash, f)
	} else if f, ok := d.GetDictionary("F"); ok {
		File(f).Extract(output, isCommand)
	}
	if win, ok := d.GetDictionary("Win"); ok {
		File(win).Extract(output, isCommand)
	}
}
