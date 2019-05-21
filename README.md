# PDF Parser
PDF Parser is a command line tool and go library for decoding PDFs

## Installation
First, [Install Go](https://golang.org/doc/install#install)

Then install/update the PDF Parser
```bash
go get -u github.com/KarmaPenny/pdfparser
```

#### To uninstall the PDF parser run:
```bash
go clean -i github.com/KarmaPenny/pdfparser && rm -rf $(go env GOPATH)/src/github.com/KarmaPenny/pdfparser
```

## Usage
#### Command Line
```bash
$(go env GOPATH)/bin/pdfparser input.pdf > output.pdf
```

#### Library
```go
package main

import (
	"fmt"
	"github.com/KarmaPenny/pdfparser/pdf"
	"os"
)

func main() {
	// open the pdf
	parser, err := pdf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer parser.Close()

	// load the cross reference
	if err := parser.LoadXref(); err != nil {
		// try to repair a bad cross reference
		fmt.Fprintln(os.Stderr, "Bad Xref")
		parser.RepairXref()
	}

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
		fmt.Fprint(os.Stdout, object)
	}
}
```

## Unit Testing
```bash
go test github.com/KarmaPenny/pdfparser
```
