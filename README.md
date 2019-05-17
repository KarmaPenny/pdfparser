# PDF Parser
PDF Parser is a command line tool and go library for decoding PDFs

## Install
First, [Install Go](https://golang.org/doc/install#install)

Then install PDF Parser by running the following command:
```bash
go get github.com/KarmaPenny/pdfparser
```

## Update
```bash
go get -u github.com/KarmaPenny/pdfparser
```

## Uninstall
```bash
go clean -i github.com/KarmaPenny/pdfparser
rm -rf $(go env GOPATH)/src/github.com/KarmaPenny/pdfparser
```

## Command Line Usage
```bash
$(go env GOPATH)/bin/pdf_parser test.pdf > output.pdf
```

## Library Usage
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
cd $(go env GOPATH)/src/github.com/KarmaPenny/pdfparser
go test
```
