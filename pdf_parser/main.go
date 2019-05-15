package main

import (
	"github.com/KarmaPenny/pdf_parser"
	"os"
)

func main() {
	pdf_parser.Parse(os.Args[1])
}
