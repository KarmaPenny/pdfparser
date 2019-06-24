package pdf

import (
	"fmt"
	"os"
)

func Parse(file_path string, password string, output_dir string) error {
	// open the pdf
	file, err := os.Open(file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	// create output directory and files
	output, err := NewOutput(output_dir)
	defer output.Close()
	if err != nil {
		return err
	}

	// create a new parser
	parser := NewParser(file, output)

	// load the pdf
	Debug("Loading xref")
	if err := parser.Load(password); err != nil {
		return err
	}

	// extract and dump all objects
	for object_number, xref_entry := range parser.Xref {
		if xref_entry.Type == XrefTypeIndirectObject {
			Debug("Extracting object %d", object_number)
			object := parser.GetObject(object_number)
			object.Extract(output)
			fmt.Fprintln(output.Raw, object.String())
		}
	}

	return nil
}
