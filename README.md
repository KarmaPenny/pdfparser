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
	reader, err := pdf.Open(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer reader.Close()

	// print all objects in xref
	for number := range reader.Xref {
		// read object
		object, err := reader.ReadObject(number)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to read object %d: %s\n", number, err)
		}

		// print object
		object.Print()
	}
}
```

## Unit Testing
```bash
go test github.com/KarmaPenny/pdfparser
```
