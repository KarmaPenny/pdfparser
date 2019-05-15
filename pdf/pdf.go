package pdf

import (
	"bufio"
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"errors"
	"fmt"
	tiff_lzw "golang.org/x/image/tiff/lzw"
	"io"
	"math"
	"os"
	"regexp"
	"strconv"
)

var start_xref_scan_buffer_size int64 = 256
var start_xref_regexp = regexp.MustCompile(`startxref\s*(\d+)\s*%%EOF`)
var whitespace = []byte("\x00\t\n\f\r ")
var delimiters = []byte("()<>[]/%")

type Pdf struct {
	*os.File
	tokenizer *bufio.Reader
	Xref map[int64]*XrefEntry
	xref_offsets map[int64]interface{}
}

func Open(name string) (*Pdf, error) {
	// open the pdf file
	file, err := os.Open(name)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Failed to open PDF: %s", err))
	}
	return &Pdf{file, bufio.NewReader(file), map[int64]*XrefEntry{}, map[int64]interface{}{}}, nil
}

func (pdf *Pdf) LoadXref() error {
	// get the offset to the first xref table
	start_xref_offset, err := pdf.getStartXrefOffset()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to get startxref offset: %s", err))
	}

	// load the cross reference table
	err = pdf.loadXref(start_xref_offset)
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to load cross reference table: %s", err))
	}
	return nil
}

// returns the offset to the first xref table
func (pdf *Pdf) getStartXrefOffset() (int64, error) {
	// start reading from the end of the file
	offset, err := pdf.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("Unable to get file_size: %s", err))
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
			return 0, errors.New(fmt.Sprintf("Scan for startxref failed: %s", err))
		}

		// check for start xref
		matches := start_xref_regexp.FindAllSubmatch(buffer, -1)
		if matches != nil {
			// return the last most start xref offset
			return strconv.ParseInt(string(matches[len(matches)-1][1]), 10, 64)
		}
	}

	// start xref not found
	return 0, errors.New("startxref marker not found.")
}

// Seek seeks the io.Reader to offset and resets the tokenizer
func (pdf *Pdf) Seek(offset int64, whence int) (int64, error) {
	pdf.tokenizer.Reset(pdf)
	return pdf.File.Seek(offset, whence)
}

// CurrentOffset returns the current byte offset of the tokenizer
func (pdf *Pdf) CurrentOffset() (int64, error) {
	offset, err := pdf.File.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}
	return offset - int64(pdf.tokenizer.Buffered()), nil
}

func (pdf *Pdf) loadXref(offset int64) error {
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
		return errors.New(fmt.Sprintf("Xref start not found: %s", err))
	}

	// xref is a table
	if s == "xref" {
		return pdf.loadXrefTable()
	}

	// xref is an object stream
	return pdf.loadXrefStream()
}

func (pdf *Pdf) loadXrefTable() error {
	// scan until end of xref table is reached
	for {
		// find next xref subsection start
		s, err := pdf.NextString()
		if err != nil {
			return errors.New(fmt.Sprintf("Xref subsection start not found: %s", err))
		}

		// if at the trailer start then stop reading xref table
		if s == "trailer" {
			break
		}

		// get subsection start
		subsection_start, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to parse xref subsection start: %s", err))
		}

		// find next xref subsection length
		s, err = pdf.NextString()
		if err != nil {
			return errors.New(fmt.Sprintf("Xref subsection length not found: %s", err))
		}

		// get xref subsection length
		subsection_length, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return errors.New(fmt.Sprintf("Failed to parse xref subsection length: %s", err))
		}

		// load each object in xref subsection
		for i := int64(0); i < subsection_length; i++ {
			// find xref entry byte offset
			s, err = pdf.NextString()
			if err != nil {
				return errors.New(fmt.Sprintf("Xref entry byte offset not found: %s", err))
			}
			byte_offset, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("Failed to parse xref entry byte offset: %s", err))
			}

			// find xref entry generation
			s, err = pdf.NextString()
			if err != nil {
				return errors.New(fmt.Sprintf("Xref entry generation not found: %s", err))
			}
			generation, err := strconv.ParseInt(s, 10, 64)
			if err != nil {
				return errors.New(fmt.Sprintf("Failed to parse xref entry generation: %s", err))
			}

			// find xref entry in use flag
			s, err = pdf.NextString()
			if err != nil {
				return errors.New(fmt.Sprintf("Xref entry in use flag not found: %s", err))
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
		return errors.New(fmt.Sprintf("Trailer dictionary not found: %s", err))
	}

	// load previous xref section if it exists
	if prev_xref_offset, err := pdf.GetInt64FromDictionary(trailer, "/Prev"); err == nil {
		return pdf.loadXref(prev_xref_offset)
	}

	return nil
}

func (pdf *Pdf) loadXrefStream() error {
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
		return errors.New(fmt.Sprintf("Trailer dictionary not found: %s", err))
	}

	// get the index and width arrays
	index, err := pdf.GetArrayFromDictionary(trailer, "/Index")
	if err != nil {
		return err
	}
	width, err := pdf.GetArrayFromDictionary(trailer, "/W")
	if err != nil {
		return err
	}

	// get widths of each field
	type_width, err := pdf.GetInt64FromArray(width, 0)
	if err != nil {
		return err
	}
	offset_width, err := pdf.GetInt64FromArray(width, 1)
	if err != nil {
		return err
	}
	generation_width, err := pdf.GetInt64FromArray(width, 2)
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
		subsection_start, err := pdf.GetInt64FromArray(index, i)
		if err != nil {
			return err
		}
		subsection_length, err := pdf.GetInt64FromArray(index, i + 1)
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
	if prev_xref_offset, err := pdf.GetInt64FromDictionary(trailer, "/Prev"); err == nil {
		return pdf.loadXref(prev_xref_offset)
	}

	return nil
}

func ReadInt(reader io.Reader, width int64) (v int64, err error) {
	data := make([]byte, width)
	bytes_read, err := reader.Read(data)
	if int64(bytes_read) != width && err != nil {
		return v, err
	}
	for i := 0; i < len(data); i++ {
		v *= 256
		v += int64(data[i])
	}
	return v, nil
}


func (pdf *Pdf) PrintObject(number int64) error {
	if xref_entry, ok := pdf.Xref[number]; ok && xref_entry.Type == 1 {
		// seek to start of object
		_, err := pdf.Seek(xref_entry.Offset, io.SeekStart)
		if err != nil {
			return err
		}

		// print object number, generation and start marker
		s, err := pdf.NextString()
		if err != nil {
			return err
		}
		fmt.Printf("%d ", number)
		s, err = pdf.NextString()
		if err != nil {
			return err
		}
		fmt.Printf("%d ", xref_entry.Generation)
		s, err = pdf.NextString()
		if err != nil {
			return err
		}
		fmt.Println(s)

		// get the value of the reference
		value, err := pdf.NextObject()
		if err != nil {
			return err
		}
		fmt.Println(value.String())

		// get next string
		s, err = pdf.NextString()
		if err != nil {
			return err
		}
		fmt.Println(s)

		// if this is a stream object
		if s == "stream" {
			if d, ok := value.(Dictionary); ok {
				data, err := pdf.ReadStream(d)
				if err != nil {
					fmt.Fprintf(os.Stderr, "Failed to read stream for %d at %d: %s\n", number, xref_entry.Offset, err)
				}
				fmt.Println(string(data))
				fmt.Println("endstream")
				fmt.Println("endobj")
				return nil
			}
			return errors.New("Stream is missing dictionary")
		}
	}
	return nil
}

func (pdf *Pdf) NextString() (string , error) {
	object, err := pdf.NextObject()
	if err != nil {
		return "", err
	}
	if s, ok := object.(*Token); ok {
		return s.String(), nil
	}
	return "", errors.New("Expected string.")
}

func (pdf *Pdf) NextDictionary() (Dictionary, error) {
	object, err := pdf.NextObject()
	if err != nil {
		return nil, err
	}
	if d, ok := object.(Dictionary); ok {
		return d, nil
	}
	return nil, errors.New("Expected dictionary.")
}

func (pdf *Pdf) NextObject() (fmt.Stringer, error) {
	// get next token
	token, err := pdf.NextToken()
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
		generation_token, err1 := pdf.NextToken()
		reference_token, err2 := pdf.NextToken()
		if err1 == nil && err2 == nil && generation_token.IsNumber && reference_token.String() == "R" {
			number, err := strconv.ParseInt(token.String(), 10, 64)
			if err != nil {
				return nil, err
			}
			generation, err := strconv.ParseInt(generation_token.String(), 10, 64)
			if err != nil {
				return nil, err
			}
			return NewReference(number, generation), nil
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

func (pdf *Pdf) NextToken() (*Token, error) {
	// skip leading whitespace
	b, err := pdf.SkipWhitespace()
	if err != nil {
		return nil, err
	}

	// start a new token
	token := NewToken(b)

	// if start or end of array then return as token
	if b == '[' || b == ']' {
		return token, nil
	}

	// if start of string
	if b == '(' {
		// find balanced closing bracket
		for open_parens := 1; open_parens > 0; {
			// read next byte
			b, err = pdf.tokenizer.ReadByte()
			if err != nil {
				return nil, err
			}

			// if this is the start of an escape sequence
			if b == '\\' {
				// read next byte
				b, err = pdf.tokenizer.ReadByte()
				if err != nil {
					return nil, err
				}

				// ignore line breaks \n or \r or \r\n
				if b == '\n' {
					continue
				}
				if b == '\r' {
					// read next byte
					b, err = pdf.tokenizer.ReadByte()
					if err != nil {
						return nil, err
					}
					// if byte is not a new line then unread it
					if b != '\n' {
						err = pdf.tokenizer.UnreadByte()
						if err != nil {
							return nil, err
						}
					}
					continue
				}

				// if this is the start of an octal character code
				if b >= '0' && b <= '7' {
					// add byte to character code
					code := bytes.NewBuffer([]byte{b})

					// add at most 2 more bytes to code
					for i := 0; i < 2; i++ {
						// read next byte
						b, err = pdf.tokenizer.ReadByte()
						if err != nil {
							return nil, err
						}

						// if next byte is not part of the octal code
						if b < '0' || b > '7' {
							// unread the byte and stop collecting code
							err = pdf.tokenizer.UnreadByte()
							if err != nil {
								return nil, err
							}
							break
						}

						// add byte to code
						code.WriteByte(b)
					}

					// convert code into byte
					val, err := strconv.ParseInt(string(code.Bytes()), 8, 8)
					if err != nil {
						return nil, err
					}
					b = byte(val)
				}

				// add byte to token and continue
				token.WriteByte(b)
				continue
			}

			// add byte to token
			token.WriteByte(b)

			// keep track of number of open parens
			if b == '(' {
				open_parens++
			} else if b == ')' {
				open_parens--
			}
		}

		// return string
		return token, nil
	}

	// if start of name
	if b == '/' {
		// parse name
		for {
			// read in the next byte
			b, err = pdf.tokenizer.ReadByte()
			if err != nil {
				return nil, err
			}

			// if the next byte is whitespace or delimiter then unread it and return the token
			if bytes.IndexByte(delimiters, b) >= 0 || bytes.IndexByte(whitespace, b) >= 0 {
				err = pdf.tokenizer.UnreadByte()
				if err != nil {
					return nil, err
				}
				return token, nil
			}

			// if next byte is the start of a hex character code
			if b == '#' {
				// read the next 2 bytes
				code, err := pdf.tokenizer.Peek(2)
				if err != nil {
					return nil, err
				}
				_, err = pdf.tokenizer.Discard(2)
				if err != nil {
					return nil, err
				}

				// convert the hex code to a byte
				val, err := strconv.ParseInt(string(code), 16, 8)
				if err != nil {
					return nil, err
				}
				b = byte(val)
			}

			// add byte to token
			token.WriteByte(b)
		}
	}

	// if start of hex string or dictionary
	if b == '<' {
		// get next byte
		b, err = pdf.SkipWhitespace()
		if err != nil {
			return nil, err
		}

		// if this is the dictionary start marker then return token
		if b == '<' {
			token.WriteByte(b)
			return token, nil
		}

		for {
			// get next byte
			b2, err := pdf.SkipWhitespace()
			if err != nil {
				return nil, err
			}

			// early end of hex string last character is assumed to be 0
			if b2 == '>' {
				// add decoded byte to token
				v, err := strconv.ParseUint(string([]byte{b, '0'}), 16, 8)
				if err != nil {
					return nil, err
				}
				token.WriteByte(byte(v))

				// add terminating marker to token and return
				token.WriteByte('>')
				return token, nil
			}

			// add decoded byte to token
			v, err := strconv.ParseUint(string([]byte{b, b2}), 16, 8)
			if err != nil {
				return nil, err
			}
			token.WriteByte(byte(v))

			// get next byte
			b, err = pdf.SkipWhitespace()
			if err != nil {
				return nil, err
			}

			// if end of hex string
			if b == '>' {
				// add terminating marker to token and return
				token.WriteByte('>')
				return token, nil
			}
		}
	}

	// if end of dictionary
	if b == '>' {
		// get the next byte
		b, err = pdf.tokenizer.ReadByte()
		if err != nil {
			return nil, err
		}
		token.WriteByte(b)

		// confirm token is a dictionary end
		if b != '>' {
			return nil, errors.New(fmt.Sprintf("Expected > instead of %b", b))
		}
		return token, nil
	}

	// set token is number if first byte is a digit
	token.IsNumber = b >= '0' && b <= '9'

	// ordinary token, scan until next whitespace or delimiter
	for {
		// get next byte
		b, err = pdf.tokenizer.ReadByte()
		if err != nil {
			return nil, err
		}

		// if byte is whitespace or delimiter then unread byte and return token
		if bytes.IndexByte(whitespace, b) >= 0 || bytes.IndexByte(delimiters, b) >= 0 {
			err = pdf.tokenizer.UnreadByte()
			if err != nil {
				return nil, err
			}
			return token, nil
		}

		// update is number
		token.IsNumber = token.IsNumber && b >= '0' && b <= '9'

		// add byte to token
		token.WriteByte(b)
	}
}

func (pdf *Pdf) SkipWhitespace() (byte, error) {
	for {
		// get next byte
		b, err := pdf.tokenizer.ReadByte()
		if err != nil {
			return 0, err
		}

		// advance if next byte is whitespace
		if bytes.IndexByte(whitespace, b) >= 0 {
			continue
		}

		// if next byte is start of a comment then advance until next line
		if b == '%' {
			_, err = pdf.tokenizer.ReadBytes('%')
			if err != nil {
				return 0, err
			}
			continue
		}

		// next byte is neither comment or whitespace so return
		return b, nil
	}
}

func (pdf *Pdf) ReadStream(d Dictionary) ([]byte, error) {
	// read until new line
	_, err := pdf.tokenizer.ReadBytes('\n')
	if err != nil {
		return nil, err
	}

	// get length of stream from dictionary
	stream_length, err := pdf.GetInt64FromDictionary(d, "/Length")
	if err != nil {
		return nil, err
	}

	// read stream data into buffer
	stream_data := make([]byte, stream_length)
	bytes_read, err := pdf.tokenizer.Read(stream_data)
	if bytes_read == 0 && err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read stream: %s", err))
	}

	// get list of filters
	filter, err := pdf.GetObjectFromDictionary(d, "/Filter")
	if err != nil {
		// no filters applied
		return stream_data, nil
	}

	// get decode parms
	has_decode_parms := true
	decode_parms, err := pdf.GetObjectFromDictionary(d, "/DecodeParms")
	if err != nil {
		has_decode_parms = false
	}

	// if filter entry is list of filters then apply in order
	filter_list, ok := filter.(Array)
	if ok {
		var decode_parms_list Array
		if has_decode_parms {
			decode_parms_list, ok = decode_parms.(Array)
		}
		for i, f := range filter_list {
			var parms fmt.Stringer
			if len(decode_parms_list) > i {
				parms = decode_parms_list[i]
			}
			stream_data, err = pdf.DecodeStream(f.String(), stream_data, parms)
			if err != nil {
				return stream_data, errors.New(fmt.Sprintf("failed to decode stream: %s", err))
			}
		}
		return stream_data, nil
	}

	// if filter is a single filter then apply it
	stream_data, err = pdf.DecodeStream(filter.String(), stream_data, decode_parms)
	if err != nil {
		return stream_data, errors.New(fmt.Sprintf("failed to decode stream: %s", err))
	}
	return stream_data, nil
}

func (pdf *Pdf) GetIntFromDictionary(d Dictionary, key string) (int, error) {
	object, err := pdf.GetObjectFromDictionary(d, key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(object.String(), 10, 32)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func (pdf *Pdf) GetInt64FromDictionary(d Dictionary, key string) (int64, error) {
	object, err := pdf.GetObjectFromDictionary(d, key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(object.String(), 10, 64)
}

func (pdf *Pdf) GetArrayFromDictionary(d Dictionary, key string) (Array, error) {
	object, err := pdf.GetObjectFromDictionary(d, key)
	if err != nil {
		return Array{}, err
	}
	if array, ok := object.(Array); ok {
		return array, nil
	}
	return Array{}, errors.New("object is not array")
}

// GetObjectFromDictionary gets an object from a dictionary, resolving references if needed
func (pdf *Pdf) GetObjectFromDictionary(d Dictionary, key string) (fmt.Stringer, error) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return pdf.ResolveReference(reference, map[int64]interface{}{})
		}
		return object, nil
	}
	return nil, errors.New(fmt.Sprintf("dictionary missing key: %s", key))
}

func (pdf *Pdf) GetIntFromArray(a Array, index int) (int, error) {
	object, err := pdf.GetObjectFromArray(a, index)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(object.String(), 10, 32)
	if err != nil {
		return 0, err
	}
	return int(i), nil
}

func (pdf *Pdf) GetInt64FromArray(a Array, index int) (int64, error) {
	object, err := pdf.GetObjectFromArray(a, index)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(object.String(), 10, 64)
}

// GetObjectFromArray gets an object from an array, resolving references if needed
func (pdf *Pdf) GetObjectFromArray(a Array, index int) (fmt.Stringer, error) {
	if index < len(a) {
		object := a[index]
		if reference, ok := object.(*Reference); ok {
			return pdf.ResolveReference(reference, map[int64]interface{}{})
		}
		return object, nil
	}
	return nil, errors.New("index out of bounds")
}

func (pdf *Pdf) ResolveReference(reference *Reference, resolved_references map[int64]interface{}) (fmt.Stringer, error) {
	// prevent infinite loop
	if _, ok := resolved_references[reference.Number]; ok {
		return nil, errors.New("Infiinite Indirect Object Reference loop detected")
	}
	resolved_references[reference.Number] = nil

	// if xref extry exists for this object number and the object is in use
	if xref_entry, ok := pdf.Xref[reference.Number]; ok && xref_entry.Type == 1 {
		// seek to start of object
		current_offset, err := pdf.CurrentOffset()
		if err != nil {
			return nil, err
		}
		_, err = pdf.Seek(xref_entry.Offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		// return to same place after we are done
		defer pdf.Seek(current_offset, io.SeekStart)

		// skip object number, generation and start marker
		_, err = pdf.NextObject()
		if err != nil {
			return nil, err
		}
		_, err = pdf.NextObject()
		if err != nil {
			return nil, err
		}
		_, err = pdf.NextObject()
		if err != nil {
			return nil, err
		}

		// get the value of the reference
		value, err := pdf.NextObject()
		if err != nil {
			return nil, err
		}

		// Recursively resolve reference
		if ref, ok := value.(*Reference); ok {
			return pdf.ResolveReference(ref, resolved_references)
		}

		// return the value
		return value, nil
	}

	// return null keyword if referenced object does not exist
	return NewTokenString("null"), nil
}

func (pdf *Pdf) DecodeStream(filter string, data []byte, decode_parms_stringer fmt.Stringer) ([]byte, error) {
	decode_parms, _ := decode_parms_stringer.(Dictionary)
	if filter == "/FlateDecode" {
		return pdf.FlateDecode(data, decode_parms)
	} else if filter == "/LZWDecode" {
		return pdf.LZWDecode(data, decode_parms)
	} else if filter == "/DCTDecode" {
		// don't bother decoding jpegs into images just dump them as is
		return data, nil
	}

	// return error if filter is not supported
	return data, errors.New(fmt.Sprintf("Unsupported filter: %s", filter))
}

func (pdf *Pdf) FlateDecode(data []byte, decode_parms Dictionary) ([]byte, error) {
	// create zlib reader from data
	zlib_reader, err := zlib.NewReader(bytes.NewReader(data))
	if err != nil {
		return data, err
	}
	defer zlib_reader.Close()

	// decode data with zlib reader and return
	var decoded_data bytes.Buffer
	bytes_read, err := decoded_data.ReadFrom(zlib_reader)
	if bytes_read == 0 && err != nil {
		return data, errors.New(fmt.Sprintf("failed to zlib decode stream: %s", err))
	}

	// reverse predictor
	return pdf.ReversePredictor(decoded_data.Bytes(), decode_parms)
}

func (pdf *Pdf) LZWDecode(data []byte, decode_parms Dictionary) ([]byte, error) {
	// create lzw reader from data using different implementation based on early change parm
	var lzw_reader io.ReadCloser
	early_change, err := pdf.GetIntFromDictionary(decode_parms, "/EarlyChange")
	if err != nil {
		early_change = 1
	}
	if early_change == 0 {
		lzw_reader = lzw.NewReader(bytes.NewReader(data), lzw.MSB, 8)
	} else {
		lzw_reader = tiff_lzw.NewReader(bytes.NewReader(data), tiff_lzw.MSB, 8)
	}
	defer lzw_reader.Close()

	// decode data with lzw reader and return
	var decoded_data bytes.Buffer
	bytes_read, err := decoded_data.ReadFrom(lzw_reader)
	if bytes_read == 0 && err != nil {
		return data, errors.New(fmt.Sprintf("failed to lzw decode stream: %s", err))
	}

	// reverse predictor
	return pdf.ReversePredictor(decoded_data.Bytes(), decode_parms)
}

func (pdf *Pdf) ReversePredictor(data []byte, decode_parms Dictionary) ([]byte, error) {
	// get predictor parms using default when not found
	predictor, err := pdf.GetIntFromDictionary(decode_parms, "/Predictor")
	if err != nil {
		predictor = 1
	}
	bits_per_component, err := pdf.GetIntFromDictionary(decode_parms, "/BitsPerComponent")
	if err != nil {
		bits_per_component = 8
	}
	colors, err := pdf.GetIntFromDictionary(decode_parms, "/Colors")
	if err != nil {
		colors = 1
	}
	columns, err := pdf.GetIntFromDictionary(decode_parms, "/Columns")
	if err != nil {
		columns = 1
	}

	// this is really hard with non byte length components so we don't support them
	if bits_per_component != 8 {
		return data, errors.New(fmt.Sprintf("unsupported BitsPerComponent: %d", bits_per_component))
	}

	// determine sample and row widths
	sample_width := colors
	row_width := columns * sample_width

	// throw error if row width is not positive
	if row_width <= 0 {
		return data, errors.New("invalid row width")
	}

	// no predictor applied
	if predictor == 1 {
		return data, nil
	}

	// TIFF predictor
	if predictor == 2 {
		for r := 0; r < len(data); r += row_width {
			for c := 1; c < row_width && r + c < len(data); c++ {
				data[r + c] = byte((int(data[r + c]) + int(data[r + c - 1])) % 256)
			}
		}
		return data, nil
	}

	// PNG predictors
	if predictor >= 10 && predictor <= 15 {
		// allocate buffer for decoded data
		decoded_data := make([]byte, 0, len(data))

		// png row includes a algorithm tag as first byte of each row
		png_row_width := row_width + 1

		// determine the method
		method := predictor - 10

		// for each row
		for row := 0; row * png_row_width < len(data); row++ {
			// row start
			r := row * png_row_width

			// optimum predictor allows the method to change each row based on algorithm tag
			if predictor == 15 {
				method = int(data[r * png_row_width])
			}

			// apply predictors based on method
			for c := 1; c < png_row_width && r + c < len(data); c++ {
				if method == 0 {
					// no predictor
					decoded_data = append(decoded_data, data[r + c])
				} else if method == 1 {
					// sub predictor
					left := 0
					if c > 1 {
						left = int(decoded_data[(row * row_width) + c - 2])
					}
					decoded_data = append(decoded_data, byte((int(data[r + c]) + left) % 256))
				} else if method == 2 {
					// up predictor
					up := 0
					if row > 0 {
						up = int(decoded_data[((row - 1) * row_width) + c - 1])
					}
					decoded_data = append(decoded_data, byte((int(data[r + c]) + up) % 256))
				} else if method == 3 {
					// avg predictor
					left := 0
					if c > 1 {
						left = int(decoded_data[(row * row_width) + c - 2])
					}
					up := 0
					if row > 0 {
						up = int(decoded_data[((row - 1) * row_width) + c - 1])
					}
					avg := (left + up) / 2
					decoded_data = append(decoded_data, byte((int(data[r + c]) + avg) % 256))
				} else if method == 4 {
					//paeth predictor
					left := 0
					if c > 1 {
						left = int(decoded_data[(row * row_width) + c - 2])
					}
					up := 0
					if row > 0 {
						up = int(decoded_data[((row - 1) * row_width) + c - 1])
					}
					up_left := 0
					if row > 0 && c > 1 {
						up_left = int(decoded_data[((row -1) * row_width) + c - 2])
					}
					p := left + up - up_left
					p_left := int(math.Abs(float64(p - left)))
					p_up := int(math.Abs(float64(p - up)))
					p_up_left := int(math.Abs(float64(p - up_left)))
					if p_left <= p_up && p_left <= p_up_left {
						data[r + c] = byte((int(data[r + c]) + left) % 256)
					} else if p_up <= p_up_left {
						data[r + c] = byte((int(data[r + c]) + up) % 256)
					} else {
						data[r + c] = byte((int(data[r + c]) + up_left) % 256)
					}
				} else {
					return data, errors.New("unsupported PNG predictor method")
				}
			}
		}
		return decoded_data, nil
	}

	return data, errors.New(fmt.Sprintf("unsupported predictor: %d", predictor))
}
