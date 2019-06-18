package logger

import (
	"fmt"
	"io"
	"os"
)

var defaultVerbose bool = false
var Verbose *bool = &defaultVerbose
var LogWriter io.Writer = os.Stderr

func Debug(format string, a ...interface{}) {
	if *Verbose {
		if len(a) > 0 {
			fmt.Fprintln(LogWriter, fmt.Sprintf(format, a...))
		} else {
			fmt.Fprintln(LogWriter, format)
		}
	}
}

func Error(format string, a ...interface{}) {
	if len(a) > 0 {
		fmt.Fprintln(LogWriter, fmt.Sprintf(format, a...))
	} else {
		fmt.Fprintln(LogWriter, format)
	}
}
