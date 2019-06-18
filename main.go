package main

import (
	"flag"
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"github.com/KarmaPenny/pdfparser/logger"
	"os"
)

var password *string
var extract_dir *string

func main() {
	// parse cmd args
	parse_args()

	// open the pdf
	PDF, err := pdf.Open(flag.Arg(0), *password)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer PDF.Close()

	// print all indirect objects in xref
	for n, entry := range PDF.Xref {
		if entry.Type == pdf.XrefTypeIndirectObject {
			fmt.Println(PDF.ReadObject(n))
		}
	}

	PDF.ExtractFiles(*extract_dir)
}

// parse cmd args
func parse_args() {
	logger.Verbose = flag.Bool("v", false, "display verbose messages")
	password = flag.String("p", "", "encryption password")
	extract_dir = flag.String("d", "", "extraction directory")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}
	if *extract_dir == "" {
		*extract_dir = flag.Arg(0) + ".extracted"
	}
	//os.RemoveAll(*extract_dir)
	os.Mkdir(*extract_dir, 0755)
}

// print help info
func usage() {
	fmt.Fprintln(os.Stderr, "PDF Parser - decodes a pdf file")
	fmt.Fprintf(os.Stderr, "https://github.com/KarmaPenny/pdfparser\n\n")
	fmt.Fprintf(os.Stderr, "usage: pdfparser [options] file\n\n")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}
