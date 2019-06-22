# PDF Parser
PDF Parser is a command line tool and go library for analyzing PDF files. PDF Parser extracts embedded files, text, hyperlinks and javascript from PDF files. PDF Parser also logs suspicious malformatting occurrences and other abnormalities (such as unnecessary escape sequences) that are commonly used to obfuscate malicious PDFs. PDF Parser supports AES and RC4 encryption as well as the following stream filters:

* ASCIIHexDecode
* ASCII85Decode
* FlateDecode
* LZWDecode
* RunLengthDecode

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
The following command extracts the contents of input.pdf to the input.pdf.extracted directory:
```bash
$(go env GOPATH)/bin/pdfparser input.pdf
```

#### Library
The following program extracts the contents of test.pdf to the output_dir directory using "password" for the encryption password:
```go
package main

import (
	"github.com/KarmaPenny/pdfparser/pdf"
)

func main() {
	pdf.Parse("test.pdf", "password", "output_dir")
}
```

## Unit Testing
```bash
go test github.com/KarmaPenny/pdfparser
```
