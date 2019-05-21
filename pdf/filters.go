package pdf

import (
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"encoding/binary"
	"io"
	"math"
	"strconv"
	tiff_lzw "golang.org/x/image/tiff/lzw"
)

func DecodeStream(filter string, data []byte, decode_parms_object Object) ([]byte, error) {
	// do nothing if data is empty
	if len(data) == 0 {
		return data, nil
	}

	// convert decode_parms_object to dictionary
	decode_parms, ok := decode_parms_object.(Dictionary)
	if !ok {
		decode_parms = Dictionary{}
	}

	// apply hex filter
	if filter == "/ASCIIHexDecode" {
		return ASCIIHexDecode(data)
	}

	// apply ascii 85 filter
	if filter == "/ASCII85Decode" {
		return ASCII85Decode(data)
	}

	// apply run length filter
	if filter == "/RunLengthDecode" {
		return RunLengthDecode(data)
	}

	// apply zlib filter
	if filter == "/FlateDecode" {
		return FlateDecode(data, decode_parms)
	}

	// apply lzw filter
	if filter == "/LZWDecode" {
		return LZWDecode(data, decode_parms)
	}

	// filter is not supported
	return data, ErrorUnsupported
}

func ASCIIHexDecode(data []byte) ([]byte, error) {
	// allocate buffer for decoded bytes
	decoded_data := make([]byte, 0, len(data))

	// decode data
	for i := 0; i < len(data); i++ {
		// get the first byte
		b1 := data[i]

		// skip whitespace
		if bytes.IndexByte(whitespace, b1) >= 0 {
			continue
		}

		// EOD marker
		if b1 == byte('>') {
			break
		}

		// make sure it is valid character
		if !IsHex(b1) {
			return decoded_data, NewError("ASCIIHexDecode: Illegal character: %x", b1)
		}

		// get the second byte defaulting to zero
		b2 := byte('0')
		for ; i + 1 < len(data); i++ {
			// skip whitespace
			if bytes.IndexByte(whitespace, data[i+1]) >= 0 {
				continue
			}

			// EOD marker
			if data[i+1] == byte('>') {
				break
			}

			// make sure it is valid character
			if !IsHex(data[i+1]) {
				return decoded_data, NewError("ASCIIHexDecode: Illegal character: %x", data[i+1])
			}

			// set second byte
			i++
			b2 = data[i]
			break
		}

		// add decoded byte to decoded data
		val, err := strconv.ParseUint(string([]byte{b1, b2}), 16, 8)
		if err != nil {
			return decoded_data, WrapError(err, "ASCIIHexDecode: Illegal character: %x", []byte{b1, b2})
		}
		b := byte(val)
		decoded_data = append(decoded_data, b)
	}

	return decoded_data, nil
}

func IsHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

func ASCII85Decode(data []byte) ([]byte, error) {
	reader := bytes.NewReader(data)
	decoded_data := bytes.NewBuffer([]byte{})

	v := uint32(0)
	n := 0

	for {
		b, err := reader.ReadByte()
		if err != nil || b == '~' {
			// finish partial group by adding zeros
			if n > 1 {
				for m := n; m < 5; m++ {
					v *= 85
				}

				// write result in big endian order
				buff := make([]byte, 4)
				binary.BigEndian.PutUint32(buff, v)
				decoded_data.Write(buff[:n-1])
			}

			return decoded_data.Bytes(), nil
		}

		// skip whitespace
		if bytes.IndexByte(whitespace, b) >= 0 {
			continue
		}

		// handle special case
		if b == 'z' {
			if n != 0 {
				return decoded_data.Bytes(), NewError("ASCII85Decode: z character in middle of group")
			}

			// write all zeros then continue
			decoded_data.Write([]byte{0,0,0,0})
			continue
		}

		// validate byte
		if b < '!' || b > 'u' {
			return decoded_data.Bytes(), NewError("ASCII85 decode: Invalid character %x", b)
		}

		// increment counter, multiply value by 85 then add value of new byte
		n++
		v *= 85
		v += uint32(b) - 33

		if n >= 5 {
			// write result in big endian order
			buff := make([]byte, 4)
			binary.BigEndian.PutUint32(buff, v)
			decoded_data.Write(buff)

			// reset value and count
			v = 0
			n =0
		}
	}
}

func RunLengthDecode(data []byte) ([]byte, error) {
	var decoded_data bytes.Buffer
	for i := 0; i < len(data); {
		// get length byte
		length := int(data[i])

		// EOD
		if length == 128 {
			break
		} else if length < 128 {
			// length is value of byte plus one
			length++

			// increment index
			i++
			if i >= len(data) {
				break
			}

			// copy as much data as we can up to length
			if i + length > len(data) {
				decoded_data.Write(data[i:])
				break
			} else {
				decoded_data.Write(data[i:i + length])
				i += length
			}
		} else if length > 128 {
			// increment index
			i++
			if i >= len(data) {
				break
			}

			// copy byte 257 - length times
			times := 257 - length
			for n := 0; n < times; n++ {
				decoded_data.WriteByte(data[i])
			}
			i++
		}
	}
	return decoded_data.Bytes(), nil
}

func FlateDecode(data []byte, decode_parms Dictionary) ([]byte, error) {
	// create zlib reader from data
	byte_reader := bytes.NewReader(data)
	zlib_reader, err := zlib.NewReader(byte_reader)
	if err != nil {
		return data, WrapError(err, "FlateDecode")
	}
	defer zlib_reader.Close()

	// decode data with zlib reader and return
	var decoded_data bytes.Buffer
	bytes_read, err := decoded_data.ReadFrom(zlib_reader)
	if bytes_read == 0 && err != nil {
		return data, WrapError(err, "FlateDecode")
	}

	// reverse predictor
	return ReversePredictor(decoded_data.Bytes(), decode_parms)
}

func LZWDecode(data []byte, decode_parms Dictionary) ([]byte, error) {
	// create lzw reader from data using different implementation based on early change parm
	var lzw_reader io.ReadCloser
	early_change, err := decode_parms.GetInt("/EarlyChange")
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
		return data, WrapError(err, "LZWDecode")
	}

	// reverse predictor
	return ReversePredictor(decoded_data.Bytes(), decode_parms)
}

func ReversePredictor(data []byte, decode_parms Dictionary) ([]byte, error) {
	// get predictor parms using default when not found
	predictor, err := decode_parms.GetInt("/Predictor")
	if err != nil {
		predictor = 1
	}
	bits_per_component, err := decode_parms.GetInt("/BitsPerComponent")
	if err != nil {
		bits_per_component = 8
	}
	colors, err := decode_parms.GetInt("/Colors")
	if err != nil {
		colors = 1
	}
	columns, err := decode_parms.GetInt("/Columns")
	if err != nil {
		columns = 1
	}

	// this is really hard with non byte length components so we don't support them
	if bits_per_component != 8 {
		return data, NewError("BitsPerComponent not supported: %d", bits_per_component)
	}

	// determine row widths
	row_width := columns * colors

	// invalid width
	if row_width <= 0 {
		return data, NewError("Invalid predictor row width: %d", row_width)
	}

	// no predictor applied
	if predictor == 1 {
		return data, nil
	}

	// TIFF predictor
	if predictor == 2 {
		for r := 0; r * row_width < len(data); r ++ {
			for c := 1; c < columns; c++ {
				for i := 0; i < colors; i++ {
					pos := r * row_width + c * colors + i
					if pos > len(data) {
						break
					}
					data[pos] = byte((int(data[pos]) + int(data[pos - colors])) % 256)
				}
			}
		}
		return data, nil
	}

	// PNG predictors
	if predictor >= 10 && predictor <= 15 {
		// allocate buffer for decoded data
		decoded_data := make([]byte, 0, len(data))

		// png row includes an algorithm tag as first byte of each row
		d_row_width := row_width
		row_width++

		// determine the method
		method := predictor - 10

		// for each row
		for r := 0; r < len(data); r += row_width {
			// optimum predictor allows the method to change each row based on algorithm tag
			if predictor == 15 {
				method = int(data[r])
			}

			// calculate start of row in decoded buffer
			dr := (r / row_width) * (row_width - 1)

			// apply predictors based on method
			for c := 1; c < row_width && r + c < len(data); c++ {
				// calculate column in decoded buffer
				dc := c - 1

				if method == 0 {
					// no predictor
					decoded_data = append(decoded_data, data[r + c])
				} else if method == 1 {
					// sub predictor
					left := 0
					if dc > 0 {
						left = int(decoded_data[dr + dc - 1])
					}
					decoded_data = append(decoded_data, byte((int(data[r + c]) + left) % 256))
				} else if method == 2 {
					// up predictor
					up := 0
					if dr + dc - d_row_width >= 0 {
						up = int(decoded_data[dr + dc - d_row_width])
					}
					decoded_data = append(decoded_data, byte((int(data[r + c]) + up) % 256))
				} else if method == 3 {
					// avg predictor
					left := 0
					if dc > 0 {
						left = int(decoded_data[dr + dc - 1])
					}
					up := 0
					if dr + dc - d_row_width >= 0 {
						up = int(decoded_data[dr + dc - d_row_width])
					}
					avg := (left + up) / 2
					decoded_data = append(decoded_data, byte((int(data[r + c]) + avg) % 256))
				} else if method == 4 {
					//paeth predictor
					left := 0
					if dc > 0 {
						left = int(decoded_data[dr + dc - 1])
					}
					up := 0
					if dr + dc - d_row_width >= 0 {
						up = int(decoded_data[dr + dc - d_row_width])
					}
					up_left := 0
					if dr + dc - d_row_width - 1 >= 0 && dc > 0 {
						up_left = int(decoded_data[dr + dc - d_row_width - 1])
					}
					p := left + up - up_left
					p_left := int(math.Abs(float64(p - left)))
					p_up := int(math.Abs(float64(p - up)))
					p_up_left := int(math.Abs(float64(p - up_left)))
					if p_left <= p_up && p_left <= p_up_left {
						decoded_data = append(decoded_data, byte((int(data[r + c]) + left) % 256))
					} else if p_up <= p_up_left {
						decoded_data = append(decoded_data, byte((int(data[r + c]) + up) % 256))
					} else {
						decoded_data = append(decoded_data, byte((int(data[r + c]) + up_left) % 256))
					}
				} else {
					// unknown predictor, do nothing
					decoded_data = append(decoded_data, data[r + c])
				}
			}
		}
		return decoded_data, nil
	}

	// unknown predictor
	return data, NewError("Predictor not supported: %d", predictor)
}
