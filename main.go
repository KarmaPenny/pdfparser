package main

import (
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

func main() {
	// open the pdf
	parser, err := pdf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer parser.Close()

	// load the cross reference
	if err := parser.LoadXref(); err != nil {
		// try to repair a bad cross reference
		fmt.Fprintln(os.Stderr, "Bad Xref")
		parser.RepairXref()
	}

	// check if pdf is encrypted
	if parser.IsEncrypted() {
		fmt.Fprintln(os.Stderr, "Encrypted")
		os.Exit(1)
	}

	// print all objects in xref
	for number := range parser.Xref {
		// read object
		object, err := parser.ReadObject(number)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			continue
		}

		// print object
		fmt.Fprint(os.Stdout, object)
	}
}
