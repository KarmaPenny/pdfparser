package main

import (
	"fmt"
	"os"

	"github.com/KarmaPenny/pdf_parser/pdf"
)

func main() {
	pdf_file, err := pdf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open pdf: %s\n", err)
		return
	}
	defer pdf_file.Close()

	err = pdf_file.LoadXref()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load xref: %s\n", err)
		return
	}

	err = pdf_file.PrintObject(1305)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to print object: %s\n", err)
	}
}
