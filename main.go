package main

import (
	"flag"
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

var password *string

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
	password = flag.String("p", "", "set encryption password")
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
	PDF, err := pdf.Open(flag.Arg(0))
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer PDF.Close()

	// check if pdf is encrypted
	if PDF.IsEncrypted() {
		if !PDF.SetPassword(*password) {
			fmt.Fprintln(os.Stderr, "Encrypted")
			os.Exit(1)
		}
	}

	// print all indirect objects in xref
	for n, entry := range PDF.Xref {
		if entry.Type == pdf.XrefTypeIndirectObject {
			object := PDF.ReadObject(n)
			fmt.Print(object.Number)
			fmt.Print(" ")
			fmt.Print(object.Generation)
			fmt.Println(" obj")
			fmt.Println(object.Value)
			if object.Stream != nil {
				fmt.Println("stream")
				fmt.Println(string(object.Stream))
				fmt.Println("endstream")
			}
			fmt.Println("endobj")
		}
	}
}
