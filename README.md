# PDF Parser
PDF Parser is a command line tool and go library that decrypts PDF files and extracts commands, files, javascript, text and urls. PDF Parser also logs formatting errors and abnormalities that are used to obfuscate malicious PDF files.

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

#### Run unit tests with:
```bash
go test github.com/KarmaPenny/pdfparser/pdf
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

## Output
PDF parser creates the following files in the output directory:
* *commands.txt* - New line separated list of commands run by launch actions.
* *contents.html* - Text content of the PDF.
* *errors.txt* - Log of malformaties and other abnormalities.
* *files.txt* - Manifest of embedded and external files. Each line contains an MD5 hash followed by a file path separated by a colon. Embedded files are extracted to the output directory using the MD5 hash as the file name. The MD5 hash for external files is all zeros.
* *javascript.js* - The javascript from all actions in the PDF
* *raw.pdf* - A decrypted and decoded version of the PDF
* *urls.txt* - New line separated list of URLs referenced by actions
