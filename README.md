# PDF Parser Command Line Tool
PDF Parser is a command line tool for decoding PDFs

## Installation
```bash
go install github.com/KarmaPenny/pdf_parser/...
```

## Usage
To use PDF Parser run the following command, replacing test.pdf with the path to a PDF file
```bash
$(go env GOPATH)/bin/pdf_parser test.pdf > output.pdf
```
