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

	// create output dir
	os.RemoveAll(*extract_dir)
	os.MkdirAll(*extract_dir, 0755)

	// parse the pdf
	pdf.Parse(flag.Arg(0), *password, *extract_dir)
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
}

// print help info
func usage() {
	fmt.Fprintln(os.Stderr, "PDF Parser - decodes a pdf file")
	fmt.Fprintf(os.Stderr, "https://github.com/KarmaPenny/pdfparser\n\n")
	fmt.Fprintf(os.Stderr, "usage: pdfparser [options] file\n\n")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}
