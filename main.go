package main

import (
	"flag"
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

var overwrite *bool
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
	overwrite = flag.Bool("f", false, "overwrite of output directory if it already exists")
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
	// check if output directory already exists
	if _, err := os.Stat(*extract_dir); !os.IsNotExist(err) && !*overwrite {
		fmt.Printf("output directory \"%s\" already exists, use -f to overwrite\n", *extract_dir)
		return
	}

	// parse the pdf
	if err := pdf.Parse(flag.Arg(0), *password, *extract_dir); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
	}
}
