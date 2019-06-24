package pdf

import (
	"fmt"
	"os"
)

var Verbose *bool

func Debug(format string, a ...interface{}) {
	if *Verbose {
		if len(a) > 0 {
			fmt.Fprintf(os.Stderr, format, a...)
			fmt.Fprintln(os.Stderr, "")
		} else {
			fmt.Fprintln(os.Stderr, format)
		}
	}
}
