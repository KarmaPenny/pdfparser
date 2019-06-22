package pdf

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"path"
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

		// get md5 hash of the file
		hash := md5.New()
		hash.Write(file_data)
		md5sum := hex.EncodeToString(hash.Sum(nil))

		// add file name relationship to manifest
		f, err := d.GetString("F")
		if err != nil {
			f = "unknown"
		}
		fmt.Fprintf(output.Files, "%s:%s\n", md5sum, f)

		// write file data to file in extract dir
		ioutil.WriteFile(path.Join(output.Directory, md5sum), file_data, 0644)
	} else if p, err := d.GetString("P"); err == nil {
		if f, err := d.GetString("F"); err == nil {
			fmt.Fprintf(output.Commands, "%s %s\n", f, p)
		}
	} else if f, err := d.GetString("F"); err == nil {
		fmt.Fprintf(output.Files, "unknown:%s\n", f)
	}
}
