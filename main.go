package main

import (
	"flag"
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

// print help info
func usage() {
	fmt.Fprintln(os.Stderr, "PDF Parser - decodes a pdf file")
	fmt.Fprintf(os.Stderr, "https://github.com/KarmaPenny/pdfparser\n\n")
	fmt.Fprintf(os.Stderr, "usage: pdfparser [options] file\n\n")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}

// parse cmd args
func parse_args() {
	pdf.Verbose = flag.Bool("v", false, "display verbose messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}
}

func main() {
	// parse cmd args
	parse_args()

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
	for n := range parser.Xref {
		// print object n if value is not null
		object := parser.ReadObject(n)
		value := object.Value.String()
		if value == "null" {
			continue
		}
		fmt.Printf("%d %d obj\n%s\n", object.Number, object.Generation, value)
		if object.Stream != nil {
			fmt.Printf("stream\n%s\nendstream\n", string(object.Stream))
		}
		fmt.Println("endobj")
	}
}
