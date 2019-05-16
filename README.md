# PDF Parser Command Line Tool
PDF Parser is a command line tool for decoding PDFs

## Installation
First, [Install Go](https://golang.org/doc/install#install)

Then install PDF Parser by running the following commands:
```bash
go get github.com/KarmaPenny/pdf_parser
go install github.com/KarmaPenny/pdf_parser
```

## Usage
```bash
cd $(go env GOPATH)/bin
./pdf_parser test.pdf > output.pdf
```
