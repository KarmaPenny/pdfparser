package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/adamjaso/pdfparser/pdf"
)

var overwrite *bool
var password *string
var textonly *bool

func usage() {
	fmt.Fprintln(os.Stderr, "PDF Parser - Decrypts a PDF file and extracts contents")
	fmt.Fprintln(os.Stderr, "<https://github.com/KarmaPenny/pdfparser>")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Usage: pdfparser [OPTION]... [FILE] [DIRECTORY]")
	fmt.Fprintln(os.Stderr, "Example: pdfparser -v -f -p password input.pdf output/")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Options:")
	fmt.Fprintln(os.Stderr, "  -f        overwrite output directory")
	fmt.Fprintln(os.Stderr, "  -p        decryption password")
	fmt.Fprintln(os.Stderr, "  -v        display verbose messages")
	fmt.Fprintln(os.Stderr, "  -t        output the text only to standard output")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Copyright (C) 2019 Cole Robinette")
	fmt.Fprintln(os.Stderr, "This program is free to use, redistribute, and modify under")
	fmt.Fprintln(os.Stderr, "the terms of the GNU General Public License version 3. This")
	fmt.Fprintln(os.Stderr, "program is distributed without any warranty.")
	fmt.Fprintln(os.Stderr, "<https://www.gnu.org/licenses/>")
}

func init() {
	overwrite = flag.Bool("f", false, "overwrite of output directory if it already exists")
	password = flag.String("p", "", "encryption password (default: empty)")
	textonly = flag.Bool("t", false, "output the contents.txt only")
	pdf.Verbose = flag.Bool("v", false, "display verbose messages")
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 2 {
		usage()
		os.Exit(1)
	}
}

func main() {
	if *textonly {
		bo := pdf.NewBufferOutput()
		err := pdf.ParseWithOptions(pdf.ParseOptions{
			Filename:  flag.Arg(0),
			Password:  *password,
			OutputDir: flag.Arg(1),
			Output:    bo,
		})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		} else {
			_, _ = io.Copy(os.Stdout, bo.Text)
		}
	} else {
		// check if output directory already exists
		if _, err := os.Stat(flag.Arg(1)); !os.IsNotExist(err) && !*overwrite {
			fmt.Printf("output directory \"%s\" already exists, use -f to overwrite\n", flag.Arg(1))
			return
		}

		// parse the pdf
		if err := pdf.Parse(flag.Arg(0), *password, flag.Arg(1)); err != nil {
			fmt.Fprintln(os.Stderr, err.Error())
		}
	}
}
