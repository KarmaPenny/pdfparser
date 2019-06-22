package pdf

import (
	"os"
	"path"
)

type Output struct {
	Commands *os.File
	Directory string
	Errors *os.File
	Files *os.File
	Javascript *os.File
	Raw *os.File
	Text *os.File
	URLs *os.File
}

func NewOutput(directory string) (output *Output, err error) {
	// create new output object
	output = &Output{}

	// create output dir
	os.RemoveAll(directory)
	os.MkdirAll(directory, 0755)

	// create commands file
	if output.Commands, err = os.Create(path.Join(directory, "commands.txt")); err != nil {
		return
	}

	// create errors file
	if output.Errors, err = os.Create(path.Join(directory, "errors.txt")); err != nil {
		return
	}

	// create manifest file
	if output.Files, err = os.Create(path.Join(directory, "files.txt")); err != nil {
		return
	}

	// create javascript file
	if output.Javascript, err = os.Create(path.Join(directory, "javascript.js")); err != nil {
		return
	}

	// create raw.pdf file
	if output.Raw, err = os.Create(path.Join(directory, "raw.pdf")); err != nil {
		return
	}

	// create text content file in output dir
	if output.Text, err = os.Create(path.Join(directory, "contents.txt")); err != nil {
		return
	}

	// create urls file
	if output.URLs, err = os.Create(path.Join(directory, "urls.txt")); err != nil {
		return
	}
	return
}

func (output *Output) Close() {
	if output.Commands != nil {
		output.Commands.Close()
	}
	if output.Errors != nil {
		output.Files.Close()
	}
	if output.Files != nil {
		output.Files.Close()
	}
	if output.Javascript != nil {
		output.Javascript.Close()
	}
	if output.Raw != nil {
		output.Raw.Close()
	}
	if output.Text != nil {
		output.Text.Close()
	}
	if output.URLs != nil {
		output.URLs.Close()
	}
}
