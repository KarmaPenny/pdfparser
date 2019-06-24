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

	// parse the pdf
	err := pdf.Parse(flag.Arg(0), *password, *extract_dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}

// parse cmd args
func parse_args() {
	logger.Verbose = flag.Bool("v", false, "display verbose messages")
	password = flag.String("p", "", "encryption password (default: empty)")
	extract_dir = flag.String("o", "", "output directory (default: [input.pdf].extracted)")
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
	fmt.Fprintf(os.Stderr, "usage: pdfparser [options] input.pdf\n\n")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}
