package pdf

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var start_xref_scan_buffer_size int64 = 256
var start_xref_regexp = regexp.MustCompile(`startxref\s*(\d+)\s*%%EOF`)
var start_obj_regexp = regexp.MustCompile(`\d+([\s\x00]|(%[^\n]*\n))+\d+([\s\x00]|(%[^\n]*\n))+obj`)
var whitespace = []byte("\x00\t\n\f\r ")
var delimiters = []byte("()<>[]/%")

type Pdf struct {
	*bufio.Reader
	file *os.File
	Xref map[int]*XrefEntry
	xref_offsets map[int64]interface{}
	trailer Dictionary
	encryption_key []byte
	security_handler *SecurityHandler
	xref_streams map[int]interface{}
}

func Open(path string, password string) (*Pdf, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	pdf := &Pdf{bufio.NewReader(file), file, map[int]*XrefEntry{}, map[int64]interface{}{}, Dictionary{}, []byte{}, nil, map[int]interface{}{}}

	// find the start xref offset and load the xref
	if start_xref_offset, err := pdf.getStartXrefOffset(); err == nil {
		if err = pdf.loadXref(start_xref_offset); err == nil {
			if err = pdf.IsXrefValid(); err == nil {
				if pdf.IsEncrypted() && !pdf.SetPassword(password) {
					pdf.Close()
					return pdf, ErrorPassword
				}
				return pdf, nil
			}
		}
	}

	// failed to load xref, attempt to repair it
	pdf.RepairXref()
	return pdf, nil
}

func (pdf *Pdf) Close() {
	pdf.file.Close()
}

func (pdf *Pdf) Seek(offset int64, whence int) int64 {
	pdf.Reset(pdf.file)
	new_offset, _ := pdf.file.Seek(offset, whence)
	return new_offset
}

func (pdf *Pdf) CurrentOffset() int64 {
	offset, err := pdf.file.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0
	}
	return offset - int64(pdf.Buffered())
}

func (pdf *Pdf) IsEncrypted() bool {
	return pdf.trailer.Contains("Encrypt")
}

func (pdf *Pdf) SetPassword(password string) bool {
	sh, err := NewSecurityHandler([]byte(password), pdf.trailer)
	if err != nil {
		return false
	}
	pdf.security_handler = sh
	return true
}

// getStartXrefOffset returns the offset to the first xref table
func (pdf *Pdf) getStartXrefOffset() (int64, error) {
	// start reading from the end of the file
	offset := pdf.Seek(0, io.SeekEnd)

	// read last several bytes and look for the start xref marker
	offset -= start_xref_scan_buffer_size
	if offset < 0 {
		offset = 0
	}

	// read in buffer at offset
	buffer := make([]byte, start_xref_scan_buffer_size)
	pdf.file.ReadAt(buffer, offset)

	// check for start xref
	matches := start_xref_regexp.FindAllSubmatch(buffer, -1)
	if matches != nil {
		// return the last most start xref offset
		start_xref_offset, err := strconv.ParseInt(string(matches[len(matches)-1][1]), 10, 64)
		if err != nil {
			return 0, WrapError(err, "Start xref offset is not int64: %s", string(matches[len(matches)-1][1]))
		}
		return start_xref_offset, nil
	}

	// start xref not found
	return 0, NewError("Start xref marker not found")
}

// loadXref loads an xref section starting at offset into pdf.Xref
func (pdf *Pdf) loadXref(offset int64) error {
	// track loaded xref offsets to prevent infinite loop
	if _, ok := pdf.xref_offsets[offset]; ok {
		// xref already loaded
		return nil
	}
	pdf.xref_offsets[offset] = nil

	// start tokenizing at offset
	pdf.Seek(offset, io.SeekStart)

	// if xref is a stream
	if n, err := pdf.readInt(); err == nil {
		pdf.xref_streams[n] = nil
		return pdf.readXrefStream()
	}

	// if xref is a table
	if keyword, err := pdf.readKeyword(); err == nil && keyword == KEYWORD_XREF {
		return pdf.readXrefTable()
	}

	return NewError("Expected xref table or stream")
}

// readXrefTable reads an xref table into pdf.Xref
func (pdf *Pdf) readXrefTable() error {
	// scan until end of xref table is reached
	for {
		// get subsection start
		subsection_start, err := pdf.readInt()
		if err != nil {
			// we are at the trailer
			if keyword, err := pdf.readKeyword(); err == nil && keyword == KEYWORD_TRAILER {
				break
			}
			return NewError("Expected int or trailer keyword")
		}

		// get subsection length
		subsection_length, err := pdf.readInt()
		if err != nil {
			return err
		}

		// load each object in xref subsection
		for i := 0; i < subsection_length; i++ {
			// find xref entry offset
			offset, err := pdf.readInt64()
			if err != nil {
				return err
			}

			// find xref entry generation
			generation, err := pdf.readInt()
			if err != nil {
				return err
			}

			// find xref entry in use flag
			flag, err := pdf.readKeyword()
			if err != nil {
				return err
			}
			xref_type := XrefTypeFreeObject
			if flag == KEYWORD_N {
				xref_type = XrefTypeIndirectObject
			}

			// determine object number from subsection start
			object_number := subsection_start + i

			// add the object if it is not in the xref table or the generation is higher
			if xref_entry, ok := pdf.Xref[object_number]; !ok || generation > xref_entry.Generation {
				pdf.Xref[object_number] = NewXrefEntry(offset, generation, xref_type)
			}
		}
	}

	// read in trailer dictionary
	trailer, err := pdf.readDictionary(noFilter)
	if err != nil {
		return err
	}

	// merge trailer
	for key, value := range trailer {
		if _, ok := pdf.trailer[key]; !ok {
			pdf.trailer[key] = value
		}
	}

	// load previous xref section if it exists
	if prev_xref_offset, err := trailer.GetInt64("Prev"); err == nil {
		return pdf.loadXref(prev_xref_offset)
	}

	return nil
}

// readXrefStream reads an xref stream object into pdf.Xref
func (pdf *Pdf) readXrefStream() error {
	// skip object generation and start marker
	if _, err := pdf.readInt(); err != nil {
		return err
	}
	if keyword, err := pdf.readKeyword(); err != nil || keyword != KEYWORD_OBJ {
		return NewError("Expected obj keyword")
	}

	// get the stream dictionary which is also the trailer dictionary
	trailer, err := pdf.readDictionary(noFilter)
	if err != nil {
		return err
	}

	// merge trailer
	for key, value := range trailer {
		if _, ok := pdf.trailer[key]; !ok {
			pdf.trailer[key] = value
		}
	}

	// get the index and width arrays
	index, err := trailer.GetArray("Index")
	if err != nil {
		// if there is no Index field then use default of [0 Size]
		size, err := trailer.GetNumber("Size")
		if err != nil {
			return err
		}
		index = Array{Number(0), size}
	}
	width, err := trailer.GetArray("W")
	if err != nil {
		return err
	}

	// get widths of each field
	type_width, err := width.GetInt(0)
	if err != nil {
		return err
	}
	offset_width, err := width.GetInt(1)
	if err != nil {
		return err
	}
	generation_width, err := width.GetInt(2)
	if err != nil {
		return err
	}

	// skip stream start marker
	if keyword, err := pdf.readKeyword(); err != nil || keyword != KEYWORD_STREAM {
		return NewError("Expected stream keyword")
	}

	// read in the stream data
	data := pdf.readStream(0, 0, trailer)
	data_reader := bytes.NewReader(data)

	// parse xref subsections
	for i := 0; i < len(index) - 1; i += 2 {
		// get subsection start and length
		subsection_start, err := index.GetInt(i)
		if err != nil {
			return err
		}
		subsection_length, err := index.GetInt(i + 1)
		if err != nil {
			return err
		}

		// read in each entry in subsection
		for j := 0; j < subsection_length; j++ {
			xref_type, err := ReadInt(data_reader, type_width)
			if err != nil {
				return err
			}
			offset, err := ReadInt64(data_reader, offset_width)
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
				pdf.Xref[object_number] = NewXrefEntry(offset, generation, xref_type)
			}
		}
	}

	// load previous xref section if it exists
	if prev_xref_offset, err := trailer.GetInt64("Prev"); err == nil {
		return pdf.loadXref(prev_xref_offset)
	}

	return nil
}

// IsXrefValid return true if the loaded xref data actually points to objects
func (pdf *Pdf) IsXrefValid() error {
	for _, entry := range pdf.Xref {
		if entry.Type == XrefTypeIndirectObject {
			// seek to start of object
			pdf.Seek(entry.Offset, io.SeekStart)

			// check for object number, generation and start marker
			if _, err := pdf.readInt(); err != nil {
				return err
			}
			if _, err := pdf.readInt(); err != nil {
				return err
			}
			if keyword, err := pdf.readKeyword(); err != nil || keyword != KEYWORD_OBJ {
				return NewError("Expected obj keyword")
			}
		}
	}
	return nil
}

// RepairXref attempts to rebuild the xref table by locating all obj start markers in the pdf file
func (pdf *Pdf) RepairXref() error {
	Debug("repairing xref")

	// clear the xref
	pdf.Xref = map[int]*XrefEntry{}

	// jump to start of file
	offset := pdf.Seek(0, io.SeekStart)

	for {
		// scan for object start marker
		index := start_obj_regexp.FindReaderIndex(pdf)
		if index == nil {
			break
		}

		// seek to start of object
		pdf.Seek(offset + int64(index[0]), io.SeekStart)

		// get object number, generation
		n, err := pdf.readInt();
		if err != nil {
			return err
		}
		g, err := pdf.readInt();
		if err != nil {
			return err
		}

		// add xref entry
		pdf.Xref[n] = NewXrefEntry(offset + int64(index[0]), g, XrefTypeIndirectObject)

		// seek to end of object start marker
		offset = pdf.Seek(offset + int64(index[1]), io.SeekStart)
	}

	Debug("repaired")
	Debug("loaded %d xref entries", len(pdf.Xref))
	return nil
}

func (pdf *Pdf) ReadObject(number int) *IndirectObject {
	Debug("Reading object %d", number)

	// create a new indirect object
	object := NewIndirectObject(number)

	if xref_entry, ok := pdf.Xref[number]; ok {
		// set the generation number
		object.Generation = xref_entry.Generation

		// if object is in use
		if xref_entry.Type == XrefTypeIndirectObject {
			// seek to start of object
			pdf.Seek(xref_entry.Offset, io.SeekStart)

			// skip object number, generation and start marker
			pdf.readInt()
			pdf.readInt()
			pdf.readKeyword()

			// initialize string decryption filter
			var string_filter CryptFilter = noFilter
			if pdf.security_handler != nil {
				string_filter = pdf.security_handler.string_filter
			}
			string_filter = string_filter.Init(number, xref_entry.Generation)

			// get the value of the object
			Debug("Reading object value")
			object.Value, _ = pdf.readObject(string_filter)

			// get next keyword
			if keyword, err := pdf.readKeyword(); err == nil && keyword == KEYWORD_STREAM {
				Debug("Reading object stream")
				// get stream dictionary
				d, ok := object.Value.(Dictionary)
				if !ok {
					d = Dictionary{}
				}

				// read the stream
				object.Stream = pdf.readStream(number, xref_entry.Generation, d)
			}
		}
	}

	Debug("Done")
	return object
}

func (pdf *Pdf) readStream(n int, g int, d Dictionary) []byte {
	// create buffers for stream data
	stream_data := bytes.NewBuffer([]byte{})

	// read until new line
	for {
		b, err := pdf.ReadByte()
		if err != nil {
			return stream_data.Bytes()
		}

		// if new line then we are at the start of the stream data
		if b == '\n' {
			break
		}

		// if carriage return check if next byte is line feed
		if b == '\r' {
			b, err := pdf.ReadByte()
			if err != nil {
				return stream_data.Bytes()
			}
			// if not new line then put it back cause it is part of the stream data
			if b != '\n' {
				pdf.UnreadByte()
			}
			break
		}
	}

	// read first 9 bytes to get started
	end_buff := bytes.NewBuffer([]byte{})
	buff := make([]byte, 9)
	bytes_read, _ := pdf.Read(buff)
	if bytes_read > 0 {
		end_buff.Write(buff[:bytes_read])
	}

	// read in stream data until endstream marker
	for {
		if end_buff.String() == "endstream" {
			// truncate last new line from stream_data and stop reading stream data
			l := stream_data.Len()
			if l-1 >= 0 && stream_data.Bytes()[l-1] == '\n' {
				if l-2 >= 0 && stream_data.Bytes()[l-2] == '\r' {
					stream_data.Truncate(l-2)
				} else {
					stream_data.Truncate(l-1)
				}
			} else if l-1 >= 0 && stream_data.Bytes()[l-1] == '\r' {
				stream_data.Truncate(l-1)
			}
			break
		}

		// add first byte of end_buff to stream_data
		b, err := end_buff.ReadByte()
		if err != nil {
			break
		}
		stream_data.WriteByte(b)

		// add next byte of stream to end_buff
		b, err = pdf.ReadByte()
		if err != nil {
			stream_data.Write(end_buff.Bytes())
			break
		}
		end_buff.WriteByte(b)
	}

	// get stream_data_bytes
	stream_data_bytes := stream_data.Bytes()

	// create list of filters
	filter_list, err := d.GetArray("Filter")
	if err != nil {
		if filter, err := d.GetName("Filter"); err == nil {
			filter_list = Array{Name(filter)}
		} else {
			filter_list = Array{}
		}
	}

	// create list of decode parms
	decode_parms_list, err := d.GetArray("DecodeParms")
	if err != nil {
		if decode_parms, err := d.GetDictionary("DecodeParms"); err == nil {
			decode_parms_list = Array{decode_parms}
		} else {
			decode_parms_list = Array{}
		}
	}

	// decrypt stream
	var crypt_filter CryptFilter = noFilter
	if _, whitelisted := pdf.xref_streams[n]; pdf.security_handler != nil && !whitelisted {
		if t, _ := d.GetName("Type"); t == "EmbeddedFile" {
			crypt_filter = pdf.security_handler.file_filter
		} else {
			crypt_filter = pdf.security_handler.stream_filter
		}
		if len(filter_list) > 0 {
			if filter, _ := filter_list.GetName(0); filter == "Crypt" {
				decode_parms, _ := decode_parms_list.GetDictionary(0)
				filter_name, err := decode_parms.GetName("Name")
				if err != nil {
					filter_name = "Identity"
				}
				if cf, exists := pdf.security_handler.crypt_filters[filter_name]; exists {
					crypt_filter = cf
				}
				filter_list = filter_list[1:]
				if len(decode_parms_list) > 0 {
					decode_parms_list = decode_parms_list[1:]
				}
			}
		}
	}
	crypt_filter = crypt_filter.Init(n, g)
	stream_data_bytes = crypt_filter.Decrypt(stream_data_bytes)

	// decode stream
	for i := 0; i < len(filter_list); i++ {
		filter, _ := filter_list.GetName(i)
		decode_parms, _ := decode_parms_list.GetDictionary(i)
		stream_data_bytes, err = DecodeStream(filter, stream_data_bytes, decode_parms)
		if err != nil {
			// stop when decode error enountered
			Debug("failed to decode stream: %s", err)
			return stream_data_bytes
		}
	}
	return stream_data_bytes
}

func (pdf *Pdf) readObject(string_filter CryptFilter) (Object, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// peek at next 2 bytes to determine object type
	b, _ := pdf.Peek(2)
	if len(b) == 0 {
		return KEYWORD_NULL, ErrorRead
	}

	// handle names
	if b[0] == '/' {
		return pdf.readName()
	}

	// handle arrays
	if b[0] == '[' {
		return pdf.readArray(string_filter)
	}
	if b[0] == ']' {
		pdf.Discard(1)
		return KEYWORD_NULL, EndOfArray
	}

	// handle strings
	if b[0] == '(' {
		s, err := pdf.readString()
		s = String(string_filter.Decrypt([]byte(s)))
		return s, err
	}

	// handle dictionaries
	if string(b) == "<<" {
		return pdf.readDictionary(string_filter)
	}
	if string(b) == ">>" {
		pdf.Discard(2)
		return KEYWORD_NULL, EndOfDictionary
	}

	// handle hex strings
	if b[0] == '<' {
		s, err := pdf.readHexString()
		s = String(string_filter.Decrypt([]byte(s)))
		return s, err
	}

	// handle keywords
	if (b[0] >= 'a' && b[0] <= 'z') || b[0] == 'R' {
		return pdf.readKeyword()
	}

	// handle numbers and references
	if (b[0] >= '0' && b[0] <= '9') || b[0] == '+' || b[0] == '-' || b[0] == '.' {
		number, err := pdf.readNumber()
		if err != nil {
			return number, err
		}

		// save offset so we can revert if this is not a reference
		offset := pdf.CurrentOffset()

		// if generation number does not follow then revert to saved offset and return number
		generation, err := pdf.readInt()
		if err != nil {
			pdf.Seek(offset, io.SeekStart)
			return number, nil
		}

		// if not a reference then revert to saved offset and return the number
		if keyword, err := pdf.readKeyword(); err != nil || keyword != KEYWORD_R {
			pdf.Seek(offset, io.SeekStart)
			return number, nil
		}

		// return the reference
		return NewReference(pdf, int(number), generation), nil
	}

	// report unknown object
	return KEYWORD_NULL, NewError("Expected array, dictionary, keyword, name, number, reference or string")
}

func (pdf *Pdf) readArray(string_filter CryptFilter) (Array, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// create a new array
	array := Array{}

	// read start of array marker
	b, err := pdf.ReadByte()
	if err != nil {
		return array, ErrorRead
	}
	if b != '[' {
		return array, NewError("Expected [")
	}

	// read in elements and append to array
	for {
		element, err := pdf.readObject(string_filter)
		if err == ErrorRead || err == EndOfArray {
			// stop if at eof or end of array
			break
		}
		array = append(array, element)
	}

	// return array
	return array, nil
}

func (pdf *Pdf) readDictionary(string_filter CryptFilter) (Dictionary, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// create new dictionary
	dictionary := Dictionary{}

	// read start of dictionary markers
	b := make([]byte, 2)
	_, err := pdf.Read(b)
	if err != nil {
		return dictionary, ErrorRead
	}
	if string(b) != "<<" {
		return dictionary, NewError("Expected <<")
	}

	// parse all key value pairs
	for {
		// read next object
		name, err := pdf.readObject(string_filter)
		if err == ErrorRead || err == EndOfDictionary {
			break
		}
		key, isName := name.(Name)
		if !isName || err != nil {
			continue
		}

		// get value
		value, err := pdf.readObject(string_filter)

		// add key value pair to dictionary
		dictionary[string(key)] = value

		// stop if at eof or end of dictionary
		if err == ErrorRead || err == EndOfDictionary {
			break
		}
	}
	return dictionary, nil
}

func (pdf *Pdf) readHexString() (String, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// create new string builder
	var s strings.Builder

	// read start of hex string marker
	b, err := pdf.ReadByte()
	if err != nil {
		return String(s.String()), ErrorRead
	}
	if b != '<' {
		return String(s.String()), NewError("Expected <")
	}

	// read hex code pairs until end of hex string or file
	for {
		code := []byte{'0', '0'}
		for i := 0; i < 2; {
			pdf.ConsumeWhitespace()
			b, err := pdf.ReadByte()
			if err != nil || b == '>' {
				if i > 0 {
					val, _ := strconv.ParseUint(string(code), 16, 8)
					s.WriteByte(byte(val))
				}
				return String(s.String()), nil
			}
			if !IsHex(b) {
				continue
			}
			code[i] = b
			i++
		}
		val, _ := strconv.ParseUint(string(code), 16, 8)
		s.WriteByte(byte(val))
	}
}

func (pdf *Pdf) readInt() (int, error) {
	value, err := pdf.readInt64()
	return int(value), err
}

func (pdf *Pdf) readInt64() (int64, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// create a new number object
	value := int64(0)

	// ensure first byte is a digit
	b, err := pdf.ReadByte()
	if err != nil || b < '0' || b > '9' {
		pdf.UnreadByte()
		return value, NewError("Expected int")
	}

	// add digit to value
	value = value * 10 + int64(b - '0')

	// parse int part
	for {
		b, err = pdf.ReadByte()
		if err != nil {
			break
		}

		// stop if no numeric char
		if b < '0' || b > '9' {
			pdf.UnreadByte()
			break
		}

		// add digit to value
		value = value * 10 + int64(b - '0')
	}

	return value, nil
}

func (pdf *Pdf) readKeyword() (Keyword, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// build keyword
	var keyword strings.Builder

	for {
		// read in the next byte
		b, err := pdf.ReadByte()
		if err != nil {
			break
		}

		// stop if not keyword character
		if (b < 'a' || b >'z') && b != 'R' {
			pdf.UnreadByte()
			break
		}

		// add character to keyword
		keyword.WriteByte(b)
	}

	// interpret keyword value
	return NewKeyword(keyword.String())
}

func (pdf *Pdf) readName() (Name, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// build name
	var name strings.Builder

	// read start of name marker
	b, err := pdf.ReadByte()
	if err != nil {
		return Name(name.String()), ErrorRead
	}
	if b != '/' {
		return Name(name.String()), NewError("Expected /")
	}

	for {
		// read in the next byte
		b, err = pdf.ReadByte()
		if err != nil {
			return Name(name.String()), nil
		}

		// if the next byte is whitespace or delimiter then unread it and return the name
		if bytes.IndexByte(delimiters, b) >= 0 || bytes.IndexByte(whitespace, b) >= 0 {
			pdf.UnreadByte()
			break
		}

		// if next byte is the start of a hex character code
		if b == '#' {
			// read in the hex code
			code := []byte{'0', '0'}
			for i := 0; i < 2; i++ {
				b, err = pdf.ReadByte()
				if err != nil {
					break
				}
				if !IsHex(b) {
					pdf.UnreadByte()
					break
				}
				code[i] = b
			}

			// convert the hex code to a byte
			val, _ := strconv.ParseUint(string(code), 16, 8)
			b = byte(val)
		}

		// add byte to name
		name.WriteByte(b)
	}

	return Name(name.String()), nil
}

func (pdf *Pdf) readNumber() (Number, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// create a new number object
	var number Number
	isReal := false
	isNegative := false

	// process first byte
	b, err := pdf.ReadByte()
	if err != nil {
		return number, ErrorRead
	}
	if b == '-' {
		isNegative = true
	} else if b >= '0' && b <= '9' {
		number = Number(float64(number) * 10 + float64(b - '0'))
	} else if b == '.' {
		isReal = true
	} else if b != '+' {
		pdf.UnreadByte()
		return number, NewError("Expected number")
	}

	// parse int part
	for !isReal {
		b, err = pdf.ReadByte()
		if err != nil {
			break
		}

		if b >= '0' && b <= '9' {
			number = Number(float64(number) * 10 + float64(b - '0'))
		} else if b == '.' {
			isReal = true
		} else {
			pdf.UnreadByte()
			break
		}
	}

	// parse real part
	if isReal {
		for i := 1; true; i++ {
			b, err = pdf.ReadByte()
			if err != nil {
				break
			}

			if b >= '0' && b <= '9' {
				number = Number(float64(number) + float64(b - '0') / (10 * float64(i)))
			} else {
				pdf.UnreadByte()
				break
			}
		}
	}

	// make negative if first byte was a minus sign
	if isNegative {
		number = -number
	}

	// return the number
	return number, nil
}

func (pdf *Pdf) readString() (String, error) {
	// consume any leading whitespace/comments
	pdf.ConsumeWhitespace()

	// create new string builder
	var s strings.Builder

	// read start of string marker
	b, err := pdf.ReadByte()
	if err != nil {
		return String(s.String()), ErrorRead
	}
	if b != '(' {
		return String(s.String()), NewError("Expected (")
	}

	// find balanced closing bracket
	for open_parens := 1; true; {
		// read next byte
		b, err = pdf.ReadByte()
		if err != nil {
			return String(s.String()), nil
		}

		// if this is the start of an escape sequence
		if b == '\\' {
			// read next byte
			b, err = pdf.ReadByte()
			if err != nil {
				s.WriteByte('\\')
				return String(s.String()), nil
			}

			// ignore escaped line breaks \n or \r or \r\n
			if b == '\n' {
				continue
			}
			if b == '\r' {
				// read next byte
				b, err = pdf.ReadByte()
				if err != nil {
					return String(s.String()), nil
				}
				// if byte is not a new line then unread it
				if b != '\n' {
					pdf.UnreadByte()
				}
				continue
			}

			// special escape values
			if b == 'n' {
				b = '\n'
			} else if b == 'r' {
				b = '\r'
			} else if b == 't' {
				b = '\t'
			} else if b == 'b' {
				b = '\b'
			} else if b == 'f' {
				b = '\f'
			}

			// if this is the start of an octal character code
			if b >= '0' && b <= '7' {
				// add byte to character code
				code := bytes.NewBuffer([]byte{b})

				// add at most 2 more bytes to code
				for i := 0; i < 2; i++ {
					// read next byte
					b, err = pdf.ReadByte()
					if err != nil {
						break
					}

					// if next byte is not part of the octal code
					if b < '0' || b > '7' {
						// unread the byte and stop collecting code
						pdf.UnreadByte()
						break
					}

					// add byte to code
					code.WriteByte(b)
				}

				// convert code into byte
				val, err := strconv.ParseUint(string(code.Bytes()), 8, 8)
				if err != nil {
					// octal code is too large so ignore last byte
					pdf.UnreadByte()
					val, _ = strconv.ParseUint(string(code.Bytes()[:code.Len()-1]), 8, 8)
				}
				b = byte(val)
			}

			// add byte to string and continue
			s.WriteByte(b)
			continue
		}

		// keep track of number of open parens
		if b == '(' {
			open_parens++
		} else if b == ')' {
			open_parens--
		}

		// stop if last paren was read
		if open_parens == 0 {
			break
		}

		// add byte to string
		s.WriteByte(b)
	}

	// return string
	return String(s.String()), nil
}

// ConsumeWhitespace reads until end of whitespace/comments
func (pdf *Pdf) ConsumeWhitespace() {
	for {
		// get next byte
		b, err := pdf.ReadByte()
		if err != nil {
			return
		}

		// consume comments and whitespace
		if b == '%' {
			pdf.ConsumeComment()
		} else if bytes.IndexByte(whitespace, b) < 0 {
			pdf.UnreadByte()
			return
		}
	}
}

func (pdf *Pdf) ConsumeComment() {
	for {
		// get next byte
		b, err := pdf.ReadByte()
		if err != nil {
			return
		}

		// stop on line feed
		if b == '\n' {
			return
		}

		// stop on carriage return
		if b == '\r' {
			// consume optional line feed
			b, err := pdf.ReadByte()
			if err != nil {
				return
			}
			if b != '\n' {
				pdf.UnreadByte()
			}
			return
		}
	}
}
