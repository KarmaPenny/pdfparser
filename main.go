package main

import (
	"fmt"
	"github.com/KarmaPenny/pdf"
	"os"
)

func main() {
	// open the pdf
	reader, err := pdf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer reader.Close()

	// print all objects in xref
	for object, xref := range reader.Xref {
		// only print objects that are in use
		if xref.Type == 1 {
			err = reader.PrintObject(object)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to print object %d: %s\n", object, err)
			}
		}
	}
}
