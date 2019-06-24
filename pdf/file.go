package pdf

import (
	"fmt"
)

var unknownHash string = "00000000000000000000000000000000"

type File Dictionary

func (file File) Extract(output *Output, isCommand bool) {
	d := Dictionary(file)

	// file specification can be a url or file
	fs, _ := d.GetString("FS")
	if fs == "URL" {
		if f, ok := d.GetString("F"); ok {
			fmt.Fprintln(output.URLs, f)
		}
	} else if ef, ok := d.GetDictionary("EF"); ok {
		// get the file data
		file_data, _ := ef.GetStream("F")

		// get the file path
		f, ok := d.GetString("F")
		if !ok {
			f = unknownHash
		}

		// dump file
		output.DumpFile(f, file_data)
	} else if p, ok := d.GetString("P"); ok {
		if f, ok := d.GetString("F"); ok {
			fmt.Fprintf(output.Files, "%s:%s\n", unknownHash, f)
			fmt.Fprintf(output.Commands, "%s %s\n", f, p)
		}
	} else if f, ok := d.GetString("F"); ok {
		if isCommand {
			fmt.Fprintf(output.Commands, "%s %s\n", f, p)
		}
		fmt.Fprintf(output.Files, "%s:%s\n", unknownHash, f)
	}
}
