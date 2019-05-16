# PDF Parser
PDF Parser is a command line tool and go library for decoding PDFs

## Installation
First, [Install Go](https://golang.org/doc/install#install)

Then install PDF Parser by running the following command:
```bash
go get github.com/KarmaPenny/pdfparser
```

## Command Line Usage
```bash
cd $(go env GOPATH)/bin
./pdf_parser test.pdf > output.pdf
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

  // print all objects in pdf
  for object := range reader.Xref {
    err = reader.PrintObject(object)
    if err != nil {
      fmt.Fprintf(os.Stderr, "Failed to print object %d: %s\n", object, err)
    }
  }
}
```
