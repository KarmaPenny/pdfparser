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
	parser, err := pdf.Open("input.pdf")
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
```

## Unit Testing
```bash
go test github.com/KarmaPenny/pdfparser
```
