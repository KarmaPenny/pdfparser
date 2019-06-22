package pdf

import (
	"fmt"
	"os"
	"path"
)

func Parse(file_path string, password string, output_dir string) error {
	// open the pdf
	file, err := os.Open(file_path)
	if err != nil {
		return err
	}
	defer file.Close()

	// create a new parser
	parser := NewParser(file)

	// load the pdf
	if err := parser.Load(password); err != nil {
		return err
	}

	// create output dir
	os.RemoveAll(output_dir)
	os.MkdirAll(output_dir, 0755)

	// create manifest file
	manifest, err := os.Create(path.Join(output_dir, "files.txt"))
	if err != nil {
		return err
	}
	defer manifest.Close()

	// create javascript file
	javascript, err := os.Create(path.Join(output_dir, "javascript.js"))
	if err != nil {
		return err
	}
	defer javascript.Close()

	// create raw.pdf file in output dir
	raw_pdf, err := os.Create(path.Join(output_dir, "raw.pdf"))
	if err != nil {
		return err
	}
	defer raw_pdf.Close()

	// create text content file in output dir
	text, err := os.Create(path.Join(output_dir, "contents.txt"))
	if err != nil {
		return err
	}
	defer text.Close()

	// create urls file
	urls, err := os.Create(path.Join(output_dir, "urls.txt"))
	if err != nil {
		return err
	}
	defer urls.Close()

	// extract and dump all objects
	for object_number, xref_entry := range parser.Xref {
		if xref_entry.Type == XrefTypeIndirectObject {
			object := parser.GetObject(object_number)
			object.Extract(manifest, javascript, text, urls, output_dir)
			fmt.Fprintln(raw_pdf, object.String())
		}
	}

	return nil
}
