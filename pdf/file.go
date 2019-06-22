package pdf

import (
	"fmt"
)

type File Dictionary

func (file File) Extract(output *Output) {
	d := Dictionary(file)

	// file specification can be a url or file
	fs, _ := d.GetString("FS")
	if string(fs) == "URL" {
		if f, err := d.GetString("F"); err == nil {
			fmt.Fprintln(output.URLs, string(f))
		}
	} else if ef, err := d.GetDictionary("EF"); err == nil {
		// get the file data
		file_data, _ := ef.GetStream("F")

		// get the file path
		f, err := d.GetString("F")
		if err != nil {
			f = "unknown"
		}

		// dump file
		output.DumpFile(f, file_data)
	} else if p, err := d.GetString("P"); err == nil {
		if f, err := d.GetString("F"); err == nil {
			fmt.Fprintf(output.Files, "unknown:%s\n", f)
			fmt.Fprintf(output.Commands, "%s %s\n", f, p)
		}
	} else if f, err := d.GetString("F"); err == nil {
		fmt.Fprintf(output.Files, "unknown:%s\n", f)
	}
}
