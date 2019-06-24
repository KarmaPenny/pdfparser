package main

import (
	"flag"
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

var password *string
var extract_dir *string

func usage() {
	fmt.Fprintln(os.Stderr, "PDF Parser - Decrypts a PDF file and extracts contents")
	fmt.Fprintf(os.Stderr, "https://github.com/KarmaPenny/pdfparser\n\n")
	fmt.Fprintf(os.Stderr, "usage: pdfparser [options] input.pdf\n\n")
	fmt.Fprintln(os.Stderr, "options:")
	flag.PrintDefaults()
}

func init() {
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

func main() {
	err := pdf.Parse(flag.Arg(0), *password, *extract_dir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
