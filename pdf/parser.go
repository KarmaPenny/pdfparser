package pdf

import (
	"bufio"
	"bytes"
	"io"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

var start_xref_scan_buffer_size int64 = 256
var start_xref_regexp = regexp.MustCompile(`startxref\s*(\d+)\s*%%EOF`)
var start_obj_regexp = regexp.MustCompile(`\d+([\s\x00]|(%[^\r\n]*))+\d+([\s\x00]|(%[^\r\n]*))+obj`)
var xref_regexp = regexp.MustCompile(`xref`)
var whitespace = []byte("\x00\t\n\f\r ")
var delimiters = []byte("()<>[]/%")

type Parser struct {
	*bufio.Reader
	seeker io.ReadSeeker
	Xref map[int]*XrefEntry
	trailer Dictionary
	security_handler *SecurityHandler
	output *Output
}

func NewParser(readSeeker io.ReadSeeker, output *Output) *Parser {
	return &Parser{bufio.NewReader(readSeeker), readSeeker, map[int]*XrefEntry{}, Dictionary{}, NewSecurityHandler(), output}
}

func (parser *Parser) Load(password string) error {
	// find location of all xref tables
	xref_offsets := parser.findXrefOffsets()

	// find location of all objects
	objects := parser.findObjects()

	// add xref stream offsets to xref offsets then sort first to last
	for _, object := range objects {
		if object.IsXrefStream {
			xref_offsets = append(xref_offsets, object.Offset)
		}
	}
	sort.Slice(xref_offsets, func(i, j int) bool { return xref_offsets[i] < xref_offsets[j] })

	// add start xref offset as last entry in xref_offsets so it overrides xref entries
	if start_xref_offset, ok := parser.getStartXrefOffset(); ok {
		xref_offsets = append(xref_offsets, start_xref_offset)
	}

	// load all xrefs
	for i := range xref_offsets {
		parser.loadXref(xref_offsets[i], map[int64]interface{}{})
	}

	// repair broken and missing xref entries
	for object_number, object := range objects {
		if xref_entry, ok := parser.Xref[object_number]; ok {
			// replace xref entry if it does not point to an object or points to the wrong object
			parser.Seek(xref_entry.Offset, io.SeekStart)
			if n, _, ok := parser.ReadObjectHeader(); ok || n != object_number {
				xref_entry.Offset = object.Offset
			}
		} else {
			// add missing object to xref
			parser.Xref[object_number] = object
		}
	}

	// setup security handler if pdf is encrypted
	if encrypt, ok := parser.trailer["Encrypt"]; ok {
		// make sure we don't decrypt the encryption dictionary
		if ref, ok := encrypt.(*Reference); ok {
			if xref_entry, ok := parser.Xref[ref.Number]; ok {
				xref_entry.IsEncrypted = false
			}
		}

		// set encryption password
		if err := parser.SetPassword(password); err != nil {
			return err
		}
	}

	return nil
}

func (parser *Parser) SetPassword(password string) error {
	return parser.security_handler.Init([]byte(password), parser.trailer)
}

// FindXrefOffsets locates all xref tables
func (parser *Parser) findXrefOffsets() []int64 {
	offsets := []int64{}

	// jump to start of file
	offset, _ := parser.Seek(0, io.SeekStart)

	for {
		// scan for xref table marker
		index := xref_regexp.FindReaderIndex(parser)
		if index == nil {
			break
		}

		// add location to offsets
		offsets = append(offsets, offset + int64(index[0]))

		// seek to end of xref marker
		offset, _ = parser.Seek(offset + int64(index[1]), io.SeekStart)
	}

	return offsets
}

// FindObjects locates all object markers
func (parser *Parser) findObjects() map[int]*XrefEntry {
	// create xref map
	objects := map[int]*XrefEntry{}

	// jump to start of file
	offset, _ := parser.Seek(0, io.SeekStart)

	for {
		// scan for object start marker
		index := start_obj_regexp.FindReaderIndex(parser)
		if index == nil {
			break
		}

		// seek to start of object
		parser.Seek(offset + int64(index[0]), io.SeekStart)

		// get object number, generation
		n, g, _ := parser.ReadObjectHeader()

		// add xref entry
		objects[n] = NewXrefEntry(offset + int64(index[0]), g, XrefTypeIndirectObject)

		// determine if object is xref stream
		d := parser.ReadDictionary(noDecryptor)
		if t, ok := d.GetName("Type"); ok && t == "XRef" {
			objects[n].IsXrefStream = true
			objects[n].IsEncrypted = false
		}

		// seek to end of object start marker
		offset, _ = parser.Seek(offset + int64(index[1]), io.SeekStart)
	}

	return objects
}

func (parser *Parser) getStartXrefOffset() (int64, bool) {
	// start reading from the end of the file
	offset, _ := parser.Seek(0, io.SeekEnd)

	// dont start past begining of file
	offset -= start_xref_scan_buffer_size
	if offset < 0 {
		offset = 0
	}

	// read in buffer at offset
	buffer := make([]byte, start_xref_scan_buffer_size)
	parser.Seek(offset, io.SeekStart)
	parser.Read(buffer)

	// check for start xref
	matches := start_xref_regexp.FindAllSubmatch(buffer, -1)
	if matches == nil {
		return 0, false
	}

	// convert match to int64
	start_xref_offset, err := strconv.ParseInt(string(matches[len(matches)-1][1]), 10, 64)
	if err != nil {
		return 0, false
	}

	// return the start xref offset
	return start_xref_offset, true
}

func (parser *Parser) loadXref(offset int64, offsets map[int64]interface{}) {
	// dont load same xref twice
	if _, ok := offsets[offset]; !ok {
		offsets[offset] = nil

		// if xref is a table
		parser.Seek(offset, io.SeekStart)
		if keyword := parser.ReadKeyword(); keyword == KEYWORD_XREF {
			parser.loadXrefTable(offsets)
		} else {
			// if xref is a stream
			parser.Seek(offset, io.SeekStart)
			if n, g, ok := parser.ReadObjectHeader(); ok {
				// prevent decrypting xref streams
				parser.Xref[n] = NewXrefEntry(offset, g, XrefTypeIndirectObject)
				parser.Xref[n].IsEncrypted = false

				// read the xref object
				parser.loadXrefStream(n, offsets)
			}
		}
	}
}

func (parser *Parser) loadXrefTable(offsets map[int64]interface{}) {
	// read all xref entries
	xrefs := map[int]*XrefEntry{}
	for {
		// get subsection start
		subsection_start, ok := parser.ReadInt()
		if !ok {
			break
		}

		// get subsection length
		subsection_length, ok := parser.ReadInt()
		if !ok {
			break
		}

		// load each object in xref subsection
		for i := 0; i < subsection_length; i++ {
			// find xref entry offset
			offset, ok := parser.ReadInt64()
			if !ok {
				break
			}

			// find xref entry generation
			generation, ok := parser.ReadInt()
			if !ok {
				break
			}

			// find xref entry in use flag
			flag := parser.ReadKeyword()
			xref_type := XrefTypeFreeObject
			if flag == KEYWORD_N {
				xref_type = XrefTypeIndirectObject
			}

			// determine object number from subsection start
			object_number := subsection_start + i

			// add the object to xrefs
			xrefs[object_number] = NewXrefEntry(offset, generation, xref_type)
		}
	}

	// read trailer keyword
	parser.ReadKeyword();

	// read in trailer dictionary
	trailer := parser.ReadDictionary(noDecryptor);

	// load previous xref section if it exists
	if prev, ok := trailer.GetInt64("Prev"); ok {
		parser.loadXref(prev, offsets)
	}

	// merge trailer
	for key, value := range trailer {
		parser.trailer[key] = value
	}

	// merge xrefs
	for key, value := range xrefs {
		parser.Xref[key] = value
	}
}

func (parser *Parser) loadXrefStream(n int, offsets map[int64]interface{}) {
	xref_stream_offset := parser.CurrentOffset()

	// Get the xref stream object
	object := parser.GetObject(n)

	// get the stream dictionary which is also the trailer dictionary
	trailer, ok := object.Value.(Dictionary)
	if !ok {
		return
	}

	// load previous xref section if it exists
	if prev, ok := trailer.GetInt64("Prev"); ok {
		parser.loadXref(prev, offsets)
	}

	// merge trailer
	for key, value := range trailer {
		parser.trailer[key] = value
	}

	// get the index and width arrays
	index, ok := trailer.GetArray("Index")
	if !ok {
		// if there is no Index field then use default of [0 Size]
		size, ok := trailer.GetNumber("Size")
		if !ok {
			return
		}
		index = Array{Number(0), size}
	}
	width, ok := trailer.GetArray("W")
	if !ok {
		return
	}

	// get widths of each field
	type_width, ok := width.GetInt(0)
	if !ok {
		return
	}
	offset_width, ok := width.GetInt(1)
	if !ok {
		return
	}
	generation_width, ok := width.GetInt(2)
	if !ok {
		return
	}

	// parse xref subsections
	data_reader := bytes.NewReader(object.Stream)
	for i := 0; i < len(index) - 1; i += 2 {
		// get subsection start and length
		subsection_start, ok := index.GetInt(i)
		if !ok {
			return
		}
		subsection_length, ok := index.GetInt(i + 1)
		if !ok {
			return
		}

		// read in each entry in subsection
		for j := 0; j < subsection_length; j++ {
			xref_type, ok := ReadInt(data_reader, type_width)
			if !ok {
				return
			}
			offset, ok := ReadInt64(data_reader, offset_width)
			if !ok {
				return
			}
			generation, ok := ReadInt(data_reader, generation_width)
			if !ok {
				return
			}

			// determine object number from subsection_start
			object_number := subsection_start + j

			// add the object to the xrefs
			parser.Xref[object_number] = NewXrefEntry(offset, generation, xref_type)
		}
	}

	// dont decrypt xref streams
	parser.Xref[object.Number] = NewXrefEntry(xref_stream_offset, object.Generation, XrefTypeIndirectObject)
	parser.Xref[object.Number].IsEncrypted = false
}

func (parser *Parser) GetObject(number int) *IndirectObject {
	object := NewIndirectObject(number)

	if xref_entry, ok := parser.Xref[number]; ok {
		if xref_entry.Type == XrefTypeIndirectObject {
			// set generation number
			object.Generation = xref_entry.Generation

			// seek to start of object
			parser.Seek(xref_entry.Offset, io.SeekStart)

			// skip object header
			parser.ReadObjectHeader()

			// initialize string decryption filter
			var string_filter CryptFilter = noFilter
			if parser.security_handler != nil && xref_entry.IsEncrypted {
				string_filter = parser.security_handler.string_filter
			}
			string_decryptor := string_filter.NewDecryptor(number, object.Generation)

			// get the value of the object
			object.Value, _ = parser.ReadObject(string_decryptor)

			// get next keyword
			if keyword := parser.ReadKeyword(); keyword == KEYWORD_STREAM {
				// get stream dictionary
				d, ok := object.Value.(Dictionary)
				if !ok {
					d = Dictionary{}
				}

				// create list of decode filters
				filter_list, ok := d.GetArray("Filter")
				if !ok {
					if filter, ok := d.GetName("Filter"); ok {
						filter_list = Array{Name(filter)}
					} else {
						filter_list = Array{}
					}
				}

				// create list of decode parms
				decode_parms_list, ok := d.GetArray("DecodeParms")
				if !ok {
					if decode_parms, ok := d.GetDictionary("DecodeParms"); ok {
						decode_parms_list = Array{decode_parms}
					} else {
						decode_parms_list = Array{}
					}
				}

				// create a stream decryptor
				var crypt_filter CryptFilter = noFilter
				if parser.security_handler != nil && xref_entry.IsEncrypted {
					// use stream filter by default
					crypt_filter = parser.security_handler.stream_filter

					// use embedded file filter if object is an embedded file
					if t, ok := d.GetName("Type"); ok && t == "EmbeddedFile" {
						crypt_filter = parser.security_handler.file_filter
					}

					// handle crypt filter override
					if len(filter_list) > 0 {
						if filter, _ := filter_list.GetName(0); filter == "Crypt" {
							decode_parms, _ := decode_parms_list.GetDictionary(0)
							filter_name, ok := decode_parms.GetName("Name")
							if !ok {
								filter_name = "Identity"
							}
							if cf, exists := parser.security_handler.crypt_filters[filter_name]; exists {
								crypt_filter = cf
							}
							filter_list = filter_list[1:]
							if len(decode_parms_list) > 0 {
								decode_parms_list = decode_parms_list[1:]
							}
						}
					}
				}
				stream_decryptor := crypt_filter.NewDecryptor(number, xref_entry.Generation)

				// read the stream
				object.Stream = parser.ReadStream(stream_decryptor, filter_list, decode_parms_list)
			}
		}
	}

	return object
}

func (parser *Parser) Seek(offset int64, whence int) (int64, error) {
	parser.Reset(parser.seeker)
	return parser.seeker.Seek(offset, whence)
}

func (parser *Parser) CurrentOffset() int64 {
	offset, err := parser.seeker.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0
	}
	return offset - int64(parser.Buffered())
}

// ReadObjectHeader reads an object header (10 0 obj) from the current position and returns the object number and generation
func (parser *Parser) ReadObjectHeader() (int, int, bool) {
	// read object number
	number, ok := parser.ReadInt()
	if !ok {
		return number, 0, false
	}

	// read object generation
	generation, ok := parser.ReadInt()
	if !ok {
		return number, generation, false
	}

	// read object start marker
	if keyword := parser.ReadKeyword(); keyword != KEYWORD_OBJ {
		return number, generation, false
	}
	return number, generation, true
}

func (parser *Parser) ReadObject(decryptor Decryptor) (Object, error) {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// peek at next 2 bytes to determine object type
	b, _ := parser.Peek(2)
	if len(b) == 0 {
		return KEYWORD_NULL, ReadError
	}

	// handle names
	if b[0] == '/' {
		return parser.ReadName(), nil
	}

	// handle arrays
	if b[0] == '[' {
		return parser.ReadArray(decryptor), nil
	}
	if b[0] == ']' {
		parser.Discard(1)
		return KEYWORD_NULL, EndOfArray
	}

	// handle strings
	if b[0] == '(' {
		return parser.ReadString(decryptor), nil
	}
	if b[0] == ')' {
		parser.Discard(1)
		return KEYWORD_NULL, EndOfString
	}

	// handle dictionaries
	if string(b) == "<<" {
		return parser.ReadDictionary(decryptor), nil
	}
	if string(b) == ">>" {
		parser.Discard(2)
		return KEYWORD_NULL, EndOfDictionary
	}

	// handle hex strings
	if b[0] == '<' {
		return parser.ReadHexString(decryptor), nil
	}
	if b[0] == '>' {
		parser.Discard(1)
		return KEYWORD_NULL, EndOfHexString
	}

	// handle numbers and references
	if (b[0] >= '0' && b[0] <= '9') || b[0] == '+' || b[0] == '-' || b[0] == '.' {
		number := parser.ReadNumber()

		// save offset so we can revert if this is not a reference
		offset := parser.CurrentOffset()

		// if generation number does not follow then revert to saved offset and return number
		generation, ok := parser.ReadInt()
		if !ok {
			parser.Seek(offset, io.SeekStart)
			return number, nil
		}

		// if not a reference then revert to saved offset and return the number
		if keyword := parser.ReadKeyword(); keyword != KEYWORD_R {
			parser.Seek(offset, io.SeekStart)
			return number, nil
		}

		// return the reference
		return NewReference(parser, int(number), generation), nil
	}

	// handle keywords
	return parser.ReadKeyword(), nil
}

func (parser *Parser) ReadArray(decryptor Decryptor) Array {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// create a new array
	array := Array{}

	// read start of array marker
	b, err := parser.ReadByte()
	if err != nil || b != '[' {
		return array
	}

	// read in elements and append to array
	for {
		element, err := parser.ReadObject(decryptor)
		if err == ReadError {
			parser.log_error(UnclosedArray)
			break
		}
		if err == EndOfArray {
			break
		}
		array = append(array, element)
	}

	// return array
	return array
}

func (parser *Parser) ReadCommand() (Keyword, Array, error) {
	// read in operands until command keyword
	operands := Array{}
	for {
		operand, err := parser.ReadObject(noDecryptor)
		if err != nil {
			return KEYWORD_NULL, operands, err
		}
		if keyword, ok := operand.(Keyword); ok {
			return keyword, operands, nil
		}
		operands = append(operands, operand)
	}
}

func (parser *Parser) ReadDictionary(decryptor Decryptor) Dictionary {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// create new dictionary
	dictionary := Dictionary{}

	// read start of dictionary markers
	b := make([]byte, 2)
	_, err := parser.Read(b)
	if err != nil || string(b) != "<<" {
		return dictionary
	}

	// parse all key value pairs
	for {
		// read next object
		name, err := parser.ReadObject(decryptor)
		if err == ReadError {
			parser.log_error(UnclosedDictionary)
			break
		}
		if err == EndOfDictionary {
			break
		}

		// skip if not a name
		key, ok := name.(Name)
		if !ok {
			parser.log_error(InvalidDictionaryKeyType)
			continue
		}

		// get value
		value, err := parser.ReadObject(decryptor)
		if err == ReadError || err == EndOfDictionary {
			parser.log_error(MissingDictionaryValue)
			break
		}

		// add key value pair to dictionary
		dictionary[string(key)] = value
	}
	return dictionary
}

func (parser *Parser) ReadHexString(decryptor Decryptor) String {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// create new string builder
	var s strings.Builder

	// read start of hex string marker
	b, err := parser.ReadByte()
	if err != nil || b != '<' {
		return String(s.String())
	}

	// read hex code pairs until end of hex string or file
	for {
		code := []byte{'0', '0'}
		for i := 0; i < 2; {
			parser.consumeWhitespace()
			b, err := parser.ReadByte()
			if err != nil || b == '>' {
				if err != nil {
					parser.log_error(UnclosedHexString)
				}
				if i > 0 {
					val, _ := strconv.ParseUint(string(code), 16, 8)
					s.WriteByte(byte(val))
				}
				s_data := []byte(s.String())
				decryptor.Decrypt(s_data)
				return String(s_data)
			}
			if !IsHex(b) {
				parser.log_error(InvalidHexStringChar)
				continue
			}
			code[i] = b
			i++
		}
		val, _ := strconv.ParseUint(string(code), 16, 8)
		s.WriteByte(byte(val))
	}
}

func (parser *Parser) ReadInt() (int, bool) {
	value, ok := parser.ReadInt64()
	return int(value), ok
}

func (parser *Parser) ReadInt64() (int64, bool) {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// create a new number object
	value := int64(0)

	// ensure first byte is a digit
	b, err := parser.ReadByte()
	if err != nil || b < '0' || b > '9' {
		parser.UnreadByte()
		return value, false
	}

	// add digit to value
	value = value * 10 + int64(b - '0')

	// parse int part
	for {
		b, err = parser.ReadByte()
		if err != nil {
			break
		}

		// stop if no numeric char
		if b < '0' || b > '9' {
			parser.UnreadByte()
			break
		}

		// add digit to value
		value = value * 10 + int64(b - '0')
	}

	return value, true
}

func (parser *Parser) ReadKeyword() Keyword {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// build keyword
	var keyword strings.Builder

	for {
		// read in the next byte
		b, err := parser.ReadByte()
		if err != nil {
			break
		}

		// stop if not keyword character
		if bytes.IndexByte(whitespace, b) >= 0 || bytes.IndexByte(delimiters, b) >= 0 {
			parser.UnreadByte()
			break
		}

		// add character to keyword
		keyword.WriteByte(b)
	}

	// interpret keyword value
	return NewKeyword(keyword.String())
}

func (parser *Parser) ReadName() Name {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// build name
	var name strings.Builder

	// read start of name marker
	b, err := parser.ReadByte()
	if err != nil || b != '/' {
		return Name(name.String())
	}

	for {
		// read in the next byte
		b, err = parser.ReadByte()
		if err != nil {
			return Name(name.String())
		}

		// if the next byte is whitespace or delimiter then unread it and return the name
		if bytes.IndexByte(delimiters, b) >= 0 || bytes.IndexByte(whitespace, b) >= 0 {
			parser.UnreadByte()
			break
		}

		// if next byte is the start of a hex character code
		if b == '#' {
			// read in the hex code
			code := []byte{'0', '0'}
			for i := 0; i < 2; i++ {
				b, err = parser.ReadByte()
				if err != nil {
					break
				}
				if !IsHex(b) {
					parser.log_error(InvalidNameEscapeChar)
					parser.UnreadByte()
					break
				}
				code[i] = b
			}

			// convert the hex code to a byte
			val, _ := strconv.ParseUint(string(code), 16, 8)
			b = byte(val)

			// if b did not need to be escaped
			if b >= '!' && b <= '~' && b != '#' && bytes.IndexByte(delimiters, b) < 0 {
				parser.log_error(UnnecessaryEscapeName)
			}
		}

		// add byte to name
		name.WriteByte(b)
	}

	return Name(name.String())
}

func (parser *Parser) ReadNumber() Number {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// create a new number object
	var number Number
	isReal := false
	isNegative := false

	// process first byte
	b, err := parser.ReadByte()
	if err != nil {
		return number
	}
	if b == '-' {
		isNegative = true
	} else if b >= '0' && b <= '9' {
		number = Number(float64(number) * 10 + float64(b - '0'))
	} else if b == '.' {
		isReal = true
	} else if b != '+' {
		parser.UnreadByte()
		return number
	}

	// parse int part
	for !isReal {
		b, err = parser.ReadByte()
		if err != nil {
			break
		}

		if b >= '0' && b <= '9' {
			number = Number(float64(number) * 10 + float64(b - '0'))
		} else if b == '.' {
			isReal = true
		} else {
			parser.UnreadByte()
			break
		}
	}

	// parse real part
	if isReal {
		for i := 1; true; i++ {
			b, err = parser.ReadByte()
			if err != nil {
				break
			}

			if b >= '0' && b <= '9' {
				number = Number(float64(number) + float64(b - '0') / (10 * float64(i)))
			} else {
				parser.UnreadByte()
				break
			}
		}
	}

	// make negative if first byte was a minus sign
	if isNegative {
		number = -number
	}

	// return the number
	return number
}

func (parser *Parser) ReadStream(decryptor Decryptor, filter_list Array, decode_parms_list Array) []byte {
	// create buffers for stream data
	stream_data := bytes.NewBuffer([]byte{})

	// read until new line
	for {
		b, err := parser.ReadByte()
		if err != nil {
			return stream_data.Bytes()
		}

		// if new line then we are at the start of the stream data
		if b == '\n' {
			break
		}

		// if carriage return check if next byte is line feed
		if b == '\r' {
			b, err := parser.ReadByte()
			if err != nil {
				return stream_data.Bytes()
			}
			// if not new line then put it back cause it is part of the stream data
			if b != '\n' {
				parser.UnreadByte()
			}
			break
		}
	}

	// read first 9 bytes to get started
	end_buff := bytes.NewBuffer([]byte{})
	buff := make([]byte, 9)
	bytes_read, _ := parser.Read(buff)
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
		b, err = parser.ReadByte()
		if err != nil {
			parser.log_error(UnclosedStream)
			stream_data.Write(end_buff.Bytes())
			break
		}
		end_buff.WriteByte(b)
	}

	// decrypt stream
	stream_data_bytes := stream_data.Bytes()
	decryptor.Decrypt(stream_data_bytes)

	// decode stream
	for i := 0; i < len(filter_list); i++ {
		filter, _ := filter_list.GetName(i)
		decode_parms, _ := decode_parms_list.GetDictionary(i)
		stream_data_bytes = DecodeStream(filter, stream_data_bytes, decode_parms)
	}

	// return the decrypted and decoded stream
	return stream_data_bytes
}

func (parser *Parser) ReadString(decryptor Decryptor) String {
	// consume any leading whitespace/comments
	parser.consumeWhitespace()

	// create new string builder
	var s strings.Builder

	// read start of string marker
	b, err := parser.ReadByte()
	if err != nil {
		return String(s.String())
	}
	if b != '(' {
		return String(s.String())
	}

	// find balanced closing bracket
	for open_parens := 1; true; {
		// read next byte
		b, err = parser.ReadByte()
		if err != nil {
			parser.log_error(UnclosedString)
			s_data := []byte(s.String())
			decryptor.Decrypt(s_data)
			return String(s_data)
		}

		// if this is the start of an escape sequence
		if b == '\\' {
			// read next byte
			b, err = parser.ReadByte()
			if err != nil {
				parser.log_error(UnclosedStringEscape)
				s.WriteByte('\\')
				s_data := []byte(s.String())
				decryptor.Decrypt(s_data)
				return String(s_data)
			}

			// ignore escaped line breaks \n or \r or \r\n
			if b == '\n' {
				continue
			}
			if b == '\r' {
				// read next byte
				b, err = parser.ReadByte()
				if err != nil {
					parser.log_error(UnclosedStringEscape)
					s_data := []byte(s.String())
					decryptor.Decrypt(s_data)
					return String(s_data)
				}
				// if byte is not a new line then unread it
				if b != '\n' {
					parser.UnreadByte()
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
					b, err = parser.ReadByte()
					if err != nil {
						parser.log_error(UnclosedStringOctal)
						break
					}

					// if next byte is not part of the octal code
					if b < '0' || b > '7' {
						// unread the byte and stop collecting code
						parser.UnreadByte()
						break
					}

					// add byte to code
					code.WriteByte(b)
				}

				// convert code into byte
				val, err := strconv.ParseUint(string(code.Bytes()), 8, 8)
				if err != nil {
					// octal code is too large so ignore last byte
					parser.log_error(InvalidOctal)
					parser.UnreadByte()
					val, _ = strconv.ParseUint(string(code.Bytes()[:code.Len()-1]), 8, 8)
				}
				b = byte(val)

				// if b did not need to be escaped
				if b >= '!' && b <= '~' && b != '\\' && b != '(' && b != ')' {
					parser.log_error(UnnecessaryEscapeString)
				}
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
	s_data := []byte(s.String())
	decryptor.Decrypt(s_data)
	return String(s_data)
}

// consumeWhitespace reads until end of whitespace/comments
func (parser *Parser) consumeWhitespace() {
	for {
		// get next byte
		b, err := parser.ReadByte()
		if err != nil {
			return
		}

		// consume comments and whitespace
		if b == '%' {
			parser.consumeComment()
		} else if bytes.IndexByte(whitespace, b) < 0 {
			parser.UnreadByte()
			return
		}
	}
}

func (parser *Parser) consumeComment() {
	for {
		// get next byte
		b, err := parser.ReadByte()
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
			b, err := parser.ReadByte()
			if err != nil {
				return
			}
			if b != '\n' {
				parser.UnreadByte()
			}
			return
		}
	}
}

func (parser *Parser) log_error(message string) {
	if parser.output != nil {
		parser.output.Error(message)
	}
}
