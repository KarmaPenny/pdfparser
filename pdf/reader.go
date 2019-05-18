package pdf

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
)

var start_xref_scan_buffer_size int64 = 256
var start_xref_regexp = regexp.MustCompile(`startxref\s*(\d+)\s*%%EOF`)

type Reader struct {
	*os.File
	tokenizer *Tokenizer
	Xref map[int64]*XrefEntry
	xref_offsets map[int64]interface{}
}

// NewPdf opens the file at path and returns a pdf object
func Open(path string) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, NewError(err)
	}
	pdf := &Reader{file, NewTokenizer(file), map[int64]*XrefEntry{}, map[int64]interface{}{}}

	// get the offset to the first xref table
	start_xref_offset, err := pdf.getStartXrefOffset()
	if err != nil {
		pdf.Close()
		return nil, NewErrXref()
	}

	// load the cross reference table
	err = pdf.loadXref(start_xref_offset)
	if err != nil {
		pdf.Close()
		return nil, NewErrXref()
	}

	return pdf, nil
}

// Seek seeks the io.Reader to offset and resets the tokenizer
func (pdf *Reader) Seek(offset int64, whence int) (int64, error) {
	pdf.tokenizer.Reset(pdf)
	new_offset, err := pdf.File.Seek(offset, whence)
	if err != nil {
		return new_offset, NewError(err)
	}
	return new_offset, nil
}

// CurrentOffset returns the current byte offset of the tokenizer
func (pdf *Reader) CurrentOffset() (int64, error) {
	offset, err := pdf.File.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, NewError(err)
	}
	return offset - int64(pdf.tokenizer.Buffered()), nil
}

// getStartXrefOffset returns the offset to the first xref table
func (pdf *Reader) getStartXrefOffset() (int64, error) {
	// start reading from the end of the file
	offset, err := pdf.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, err
	}

	// read until we find startxref marker or we reach the start of the file
	for offset >= 0 {
		// add half the buffer_size so that we do not miss markers that cross segments
		offset -= start_xref_scan_buffer_size / 2

		// prevent reading past start of file
		if offset < 0 {
			offset = 0
		}

		// read in buffer at offset
		buffer := make([]byte, start_xref_scan_buffer_size)
		_, err = pdf.ReadAt(buffer, offset)
		if err != nil && err != io.EOF {
			return 0, NewError(err)
		}

		// check for start xref
		matches := start_xref_regexp.FindAllSubmatch(buffer, -1)
		if matches != nil {
			// return the last most start xref offset
			start_xref_offset, err := strconv.ParseInt(string(matches[len(matches)-1][1]), 10, 64)
			if err != nil {
				return 0, NewError(err)
			}
			return start_xref_offset, nil
		}
	}

	// start xref not found
	return 0, NewErrorf("Start xref not found")
}

// loadXref loads an xref section starting at offset into pdf.Xref
func (pdf *Reader) loadXref(offset int64) error {
	// track loaded xref offsets to prevent infinite loop
	if _, ok := pdf.xref_offsets[offset]; ok {
		// xref already loaded
		return nil
	}
	pdf.xref_offsets[offset] = nil

	// start tokenizing at offset
	_, err := pdf.Seek(offset, io.SeekStart)
	if err != nil {
		return err
	}

	// find xref start
	s, err := pdf.NextString()
	if err != nil {
		return err
	}

	// xref is a table
	if s == "xref" {
		return pdf.readXrefTable()
	}

	// xref is an object stream
	return pdf.readXrefStream()
}

// readXrefTable reads an xref table into pdf.Xref
func (pdf *Reader) readXrefTable() error {
	// scan until end of xref table is reached
	for {
		// find next xref subsection start
		s, err := pdf.NextString()
		if err != nil {
			return err
		}

		// if at the trailer start then stop reading xref table
		if s == "trailer" {
			break
		}

		// get subsection start
		subsection_start, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return NewError(err)
		}

		// find next xref subsection length
		subsection_length, err := pdf.NextInt64()
		if err != nil {
			return err
		}

		// load each object in xref subsection
		for i := int64(0); i < subsection_length; i++ {
			// find xref entry byte offset
			s, err = pdf.NextString()
			if err != nil {
				return err
			}
			byte_offset, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return NewError(err)
			}

			// find xref entry generation
			s, err = pdf.NextString()
			if err != nil {
				return err
			}
			generation, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return NewError(err)
			}

			// find xref entry in use flag
			s, err = pdf.NextString()
			if err != nil {
				return err
			}

			// determine object number from subsection start
			object_number := subsection_start + i

			// determine object type
			type_value := int64(0)
			if s == "n" {
				type_value = int64(1)
			}

			// add the object if it is not in the xref table or the generation is higher
			if xref_entry, ok := pdf.Xref[object_number]; !ok || (ok && generation > xref_entry.Generation) {
				pdf.Xref[object_number] = NewXrefEntry(byte_offset, generation, type_value)
			}
		}
	}

	// find start of trailer dictionary
	trailer, err := pdf.NextDictionary()
	if err != nil {
		return err
	}

	// check if pdf is encrypted
	if _, err := trailer.GetObject("/Encrypt"); err == nil {
		return NewErrEncrypted()
	}

	// load previous xref section if it exists
	if prev_xref_offset, err := trailer.GetInt64("/Prev"); err == nil {
		return pdf.loadXref(prev_xref_offset)
	}

	return nil
}

// readXrefStream reads an xref stream object into pdf.Xref
func (pdf *Reader) readXrefStream() error {
	// skip generation number and obj start marker
	_, err := pdf.NextObject()
	if err != nil {
		return err
	}
	_, err = pdf.NextObject()
	if err != nil {
		return err
	}

	// get the stream dictionary which is also the trailer dictionary
	trailer, err := pdf.NextDictionary()
	if err != nil {
		return err
	}

	// check if pdf is encrypted
	if _, err := trailer.GetObject("/Encrypt"); err == nil {
		return NewErrEncrypted()
	}

	// get the index and width arrays
	index, err := trailer.GetArray("/Index")
	if err != nil {
		// if there is no Index field then use default of [0 Size]
		size, err := trailer.GetObject("/Size")
		if err != nil {
			return err
		}
		index = Array{NewTokenString("0"), size}
	}
	width, err := trailer.GetArray("/W")
	if err != nil {
		return err
	}

	// get widths of each field
	type_width, err := width.GetInt64(0)
	if err != nil {
		return err
	}
	offset_width, err := width.GetInt64(1)
	if err != nil {
		return err
	}
	generation_width, err := width.GetInt64(2)
	if err != nil {
		return err
	}

	// skip stream start marker
	_, err = pdf.NextObject()
	if err != nil {
		return err
	}

	// read in the stream data
	data, err := pdf.ReadStream(trailer)
	if err != nil {
		return err
	}
	data_reader := bytes.NewReader(data)

	// parse xref subsections
	for i := 0; i < len(index) - 1; i += 2 {
		// get subsection start and length
		subsection_start, err := index.GetInt64(i)
		if err != nil {
			return err
		}
		subsection_length, err := index.GetInt64(i + 1)
		if err != nil {
			return err
		}

		// read in each entry in subsection
		for j := int64(0); j < subsection_length; j++ {
			type_value, err := ReadInt(data_reader, type_width)
			if err != nil {
				return err
			}
			offset, err := ReadInt(data_reader, offset_width)
			if err != nil {
				return err
			}
			generation, err := ReadInt(data_reader, generation_width)
			if err != nil {
				return err
			}

			// determine object number from subsection_start
			object_number := subsection_start + j

			// add the object if it is not in the xref table or the generation is higher
			if xref_entry, ok := pdf.Xref[object_number]; !ok || (ok && generation > xref_entry.Generation) {
				pdf.Xref[object_number] = NewXrefEntry(offset, generation, type_value)
			}
		}
	}

	// load previous xref section if it exists
	if prev_xref_offset, err := trailer.GetInt64("/Prev"); err == nil {
		return pdf.loadXref(prev_xref_offset)
	}

	return nil
}

// ReadObject reads an object by looking up the number in the xref table
func (pdf *Reader) ReadObject(number int64) (*IndirectObject, error) {
	// create a new indirect object
	object := NewIndirectObject(number)

	// TODO: add support for compressed object type
	if xref_entry, ok := pdf.Xref[number]; ok && xref_entry.Type == XrefTypeIndirectObject {
		// seek to start of object
		_, err := pdf.Seek(xref_entry.Offset, io.SeekStart)
		if err != nil {
			return object, err
		}

		// get object number
		_, err = pdf.NextInt64()
		if err != nil {
			return object, err
		}

		// get generation number
		object.Generation, err = pdf.NextInt64()
		if err != nil {
			return object, err
		}

		// skip obj start marker
		_, err = pdf.NextString()
		if err != nil {
			return object, err
		}

		// get the value of the reference
		object.Value, err = pdf.NextObject()
		if err != nil {
			return object, err
		}

		// get next string
		var s string
		s, err = pdf.NextString()
		if err != nil {
			return object, err
		}

		// if this is a stream object
		if s == "stream" {
			if d, ok := object.Value.(Dictionary); ok {
				object.Stream, err = pdf.ReadStream(d)
				if err != nil {
					return object, err
				}
				return object, nil
			}
			return object, NewErrorf("Stream has no dictionary")
		}
	}
	return object, nil
}

func (pdf *Reader) ReadStream(d Dictionary) ([]byte, error) {
	// read until new line
	_, err := pdf.tokenizer.ReadBytes('\n')
	if err != nil {
		return nil, NewError(err)
	}

	// get length of stream from dictionary
	stream_length, err := d.GetInt64("/Length")
	if err != nil {
		return nil, err
	}

	// read stream data into buffer
	stream_data := make([]byte, stream_length)
	bytes_read, err := pdf.tokenizer.Read(stream_data)
	if bytes_read == 0 && err != nil {
		return nil, NewError(err)
	}

	// get list of filters
	filter, err := d.GetObject("/Filter")
	if err != nil {
		// no filters applied
		return stream_data, nil
	}

	// get decode parms
	decode_parms_object, err := d.GetObject("/DecodeParms")
	if err != nil {
		decode_parms_object = NewTokenString("null")
	}

	// if filter entry is list of filters then apply in order
	filter_list, ok := filter.(Array)
	if ok {
		decode_parms_list, ok := decode_parms_object.(Array)
		if !ok {
			decode_parms_list = Array{}
		}
		for i, f := range filter_list {
			decode_parms_list_object, err := decode_parms_list.GetObject(i)
			if err != nil {
				decode_parms_list_object = NewTokenString("null")
			}
			stream_data, err = DecodeStream(f.String(), stream_data, decode_parms_list_object)
			if err != nil {
				return stream_data, err
			}
		}
		return stream_data, nil
	}

	// if filter is a single filter then apply it
	stream_data, err = DecodeStream(filter.String(), stream_data, decode_parms_object)
	if err != nil {
		return stream_data, err
	}
	return stream_data, nil
}

func (pdf *Reader) NextInt64() (int64, error) {
	s, err := pdf.NextString()
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return value, NewError(err)
	}
	return value, nil
}

func (pdf *Reader) NextString() (string , error) {
	object, err := pdf.NextObject()
	if err != nil {
		return "", err
	}
	if s, ok := object.(*Token); ok {
		return s.String(), nil
	}
	return "", NewErrorf("Expected string")
}

func (pdf *Reader) NextDictionary() (Dictionary, error) {
	object, err := pdf.NextObject()
	if err != nil {
		return nil, err
	}
	if d, ok := object.(Dictionary); ok {
		return d, nil
	}
	return nil, NewErrorf("Expected Dictionary")
}

func (pdf *Reader) NextObject() (fmt.Stringer, error) {
	// get next token
	token, err := pdf.tokenizer.NextToken()
	if err != nil {
		return nil, err
	}

	// if the next 3 tokens form a reference
	if token.IsNumber {
		// get current offset so we can revert if need be
		current_offset, err := pdf.CurrentOffset()
		if err != nil {
			return nil, err
		}

		// check if next two tokens make a complete reference token
		generation_token, err1 := pdf.tokenizer.NextToken()
		reference_token, err2 := pdf.tokenizer.NextToken()
		if err1 == nil && err2 == nil && generation_token.IsNumber && reference_token.String() == "R" {
			number, err := strconv.ParseInt(token.String(), 10, 64)
			if err != nil {
				return nil, NewError(err)
			}
			generation, err := strconv.ParseInt(generation_token.String(), 10, 64)
			if err != nil {
				return nil, NewError(err)
			}
			return NewReference(pdf, number, generation), nil
		}

		// token is not a reference so revert tokenizer
		_, err = pdf.Seek(current_offset, io.SeekStart)
		if err != nil {
			return nil, err
		}
		return token, nil
	}

	// if token is array start
	if token.String() == "[" {
		// create new array
		array := Array{}

		// parse all elements
		for {
			// get next object
			next_object, err := pdf.NextObject()
			if err != nil {
				return nil, err
			}

			// return if next object is array end
			if t, ok := next_object.(*Token); ok && t.String() == "]" {
				return array, nil
			}

			// add next object to array
			array = append(array, next_object)
		}
	}

	// if token is dictionary start
	if token.String() == "<<" {
		// create new dictionary
		dictionary := Dictionary{}

		// parse all key value pairs
		for {
			// next object should be a key or dictionary end
			key, err := pdf.NextString()
			if err != nil {
				return nil, err
			}

			// return if key is dictionary end
			if key == ">>" {
				return dictionary, nil
			}

			// next object is the value
			value, err := pdf.NextObject()
			if err != nil {
				return nil, err
			}

			// add key value pair to dictionary
			dictionary[key] = value
		}
	}

	// return token
	return token, nil
}
