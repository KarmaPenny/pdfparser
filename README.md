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
$(go env GOPATH)/bin/pdfparser -p password input.pdf output/
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

#### commands.txt
Commands run by launch actions are logged to the commands.txt file. Example:
```
cmd.exe /c hello.exe
calc.exe
```

#### contents.txt
The text content of the PDF is written to the contents.txt file.

#### errors.txt
Format errors and other abnormailites that are sometimes used to obfuscate malicious PDF files are logged to the errors.txt file. Bellow is an example errors.txt file containing the complete list of possible log messages:
```
invalid dictionary key type
invalid hex string character
invalid name escape character
invalid octal in string
missing dictionary value
unclosed array
unclosed dictionary
unclosed hex string
unclosed stream
unclosed string
unclosed escape in string
unclosed octal in string
unnecessary espace sequence in name
unnecessary espace sequence in string
```
#### files.txt
The MD5 hash and file path of referenced embedded and external files are logged to the files.txt file. Embedded files are extracted to the output directory using the MD5 hash as the file name. The MD5 hash for external files is all zeros. Example:
```
6adb6f85e541f14d7ecec12a6af8ef65:hello.exe
00000000000000000000000000000000:C:\Windows\System32\calc.exe
```

#### javascript.js
The javascript of all actions is extracted to the javascript.js file.

#### raw.pdf
A decrypted and decoded version of the PDF is written to the raw.pdf file.

#### urls.txt
All URLs referenced by actions are extracted to the urls.txt file. Example:
```
http://www.google.com
https://github.com/KarmaPenny
```
