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
		object.Fprint(os.Stdout)
	}
}
