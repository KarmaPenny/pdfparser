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
	PDF, err := pdf.Open("test.pdf", "password")
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
}
```

## Unit Testing
```bash
go test github.com/KarmaPenny/pdfparser
```
