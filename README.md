# PDF Parser
PDF Parser is a command line tool and go library that extracts files, text, urls, commands and javascript from PDF files. PDF Parser also logs suspicious malformatting occurrences and other abnormalities (such as unnecessary escape sequences) that are commonly used to obfuscate malicious PDF files.

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
The following command extracts the contents of input.pdf to the output directory using "password" for decryption:
```bash
$(go env GOPATH)/bin/pdfparser -o output -p password input.pdf
```

#### Library
The following program extracts the contents of input.pdf to the output directory using "password" for decryption:
```go
package main

import "github.com/KarmaPenny/pdfparser/pdf"

func main() {
	pdf.Parse("input.pdf", "password", "output")
}
```

## Unit Testing
```bash
go test github.com/KarmaPenny/pdfparser
```
