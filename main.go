package main

import (
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

func main() {
	// open the pdf
	reader, err := pdf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		if stack, ok := err.(*pdf.ErrTrace); ok {
			fmt.Fprintln(os.Stderr, stack.Trace())
		}
		return
	}
	defer reader.Close()

	// print all objects in xref
	for number := range reader.Xref {
		// read object
		object, err := reader.ReadObject(number)
		if err != nil {
			if stack, ok := err.(*pdf.ErrTrace); ok {
				fmt.Fprintln(os.Stderr, err)
				fmt.Fprintln(os.Stderr, stack.Trace())
				continue
			}
		}

		// print object
		object.Print()
	}
}
