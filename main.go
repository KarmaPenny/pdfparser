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
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
	defer reader.Close()

	if reader.IsEncrypted() {
		fmt.Fprintln(os.Stderr, "Encrypted")
		os.Exit(1)
	}

	// print all objects in xref
	for number := range reader.Xref {
		// read object
		object, err := reader.ReadObject(number)
		if err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
			continue
		}

		// print object
		object.Print()
	}
}
