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
		return
	}
	defer reader.Close()

	// print all objects in xref
	for object, xref := range reader.Xref {
		// skip free objects
		if xref.Type == 0 {
			continue
		}

		// print objects that are in use
		if xref.Type == 1 {
			err = reader.PrintObject(object)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to print object %d: %s\n", object, err)
			}
		} else {
			// report objects with unsupported types
			fmt.Fprintf(os.Stderr, "Unsupported object type for object #%d: %d\n", object, xref.Type)
		}
	}
}
