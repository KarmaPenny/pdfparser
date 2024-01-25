package pdf

import (
	"fmt"
	"os"
)

type ParseOptions struct {
	Filename  string
	Password  string
	OutputDir string
	Output    *Output
}

func Parse(file_path, password, output_dir string) error {
	return ParseWithOptions(ParseOptions{
		Filename:  file_path,
		Password:  password,
		OutputDir: output_dir,
	})
}

func ParseWithOptions(opts ParseOptions) error {
	if Verbose == nil {
		Verbose = new(bool)
	}
	// open the pdf
	file, err := os.Open(opts.Filename)
	if err != nil {
		return err
	}
	defer file.Close()

	// create output directory and files
	var output *Output
	if opts.Output != nil {
		output = opts.Output
	} else {
		output, err = NewOutput(opts.OutputDir)
	}
	defer output.Close()
	if err != nil {
		return err
	}

	// create a new parser
	parser := NewParser(file, output)

	// load the pdf
	Debug("Loading xref")
	if err := parser.Load(opts.Password); err != nil {
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
