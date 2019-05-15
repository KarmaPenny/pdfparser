package pdf_parser

import (
	"fmt"
	"os"
)

func Parse(path string) {
	// open the pdf with a pdf reader
	reader, err := Open(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		return
	}
	defer reader.Close()

	// print all objects in xref
	for object_number := range reader.Xref {
		err = reader.PrintObject(object_number)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to print object %d: %s\n", object_number, err)
		}
	}
}
