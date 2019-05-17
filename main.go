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
	for number, xref := range reader.Xref {
		// skip free objects
		if xref.Type == 0 {
			continue
		}

		// report unsupported object types
		if xref.Type != 1 {
			fmt.Fprintf(os.Stderr, "Unsupported object type for object %d: %d\n", number, xref.Type)
		}

		// read object
		object, err := reader.ReadObject(number)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read object %d: %s\n", number, err)
		}

		// print object
		object.Print()
	}
}
