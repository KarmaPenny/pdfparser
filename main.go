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
	for number := range reader.Xref {
		// read object
		object, err := reader.ReadObject(number)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read object %d: %s\n", number, err)
		}

		// print object
		object.Print()
	}
}
