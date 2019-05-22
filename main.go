package main

import (
	"flag"
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

// print help info
func usage() {
	fmt.Fprintln(os.Stderr, "pdfparser - decodes a pdf file")
	fmt.Fprintf(os.Stderr, "https://github.com/KarmaPenny/pdfparser\n\n")
	fmt.Fprintf(os.Stderr, "usage: pdfparser [options] file\n\n")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}

func main() {
	// cmd args
	pdf.Verbose = flag.Bool("v", false, "display verbose log and error messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}

	// open the pdf
	parser, err := pdf.Open(flag.Arg(0))
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
