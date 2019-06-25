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

func DecodeStream(filter string, data []byte, decode_parms Dictionary) []byte {
	// do nothing if data is empty
	if len(data) == 0 {
		return data
	}

	// apply hex filter
	if filter == "ASCIIHexDecode" {
		return ASCIIHexDecode(data)
	}

	// apply ascii 85 filter
	if filter == "ASCII85Decode" {
		return ASCII85Decode(data)
	}

	// apply run length filter
	if filter == "RunLengthDecode" {
		return RunLengthDecode(data)
	}

	// apply zlib filter
	if filter == "FlateDecode" {
		return FlateDecode(data, decode_parms)
	}

	// apply lzw filter
	if filter == "LZWDecode" {
		return LZWDecode(data, decode_parms)
	}

	// filter is not supported
	return data
}

func ASCIIHexDecode(data []byte) []byte {
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

			// set second byte
			i++
			b2 = data[i]
			break
		}

		// add decoded byte to decoded data
		val, err := strconv.ParseUint(string([]byte{b1, b2}), 16, 8)
		if err != nil {
			// TODO: report illegal character
			continue
		}
		decoded_data = append(decoded_data, byte(val))
	}

	return decoded_data
}

func ASCII85Decode(data []byte) []byte {
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

			return decoded_data.Bytes()
		}

		// skip whitespace
		if bytes.IndexByte(whitespace, b) >= 0 {
			continue
		}

		// handle special case
		if b == 'z' {
			if n != 0 {
				// TODO: report invalid use of z character
				continue
			}

			// write all zeros then continue
			decoded_data.Write([]byte{0,0,0,0})
			continue
		}

		// validate byte
		if b < '!' || b > 'u' {
			// TODO: report invalid ascii85 character
			continue
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

func RunLengthDecode(data []byte) []byte {
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
	return decoded_data.Bytes()
}

func FlateDecode(data []byte, decode_parms Dictionary) []byte {
	// create zlib reader from data
	byte_reader := bytes.NewReader(data)
	zlib_reader, err := zlib.NewReader(byte_reader)
	if err != nil {
		return data
	}
	defer zlib_reader.Close()

	// decode data with zlib reader and return
	var decoded_data bytes.Buffer
	bytes_read, err := decoded_data.ReadFrom(zlib_reader)
	if bytes_read == 0 && err != nil {
		return data
	}

	// reverse predictor
	return ReversePredictor(decoded_data.Bytes(), decode_parms)
}

func LZWDecode(data []byte, decode_parms Dictionary) []byte {
	// create lzw reader from data using different implementation based on early change parm
	var lzw_reader io.ReadCloser
	early_change, ok := decode_parms.GetInt("EarlyChange")
	if !ok {
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
		return data
	}

	// reverse predictor
	return ReversePredictor(decoded_data.Bytes(), decode_parms)
}

func ReversePredictor(data []byte, decode_parms Dictionary) []byte {
	// get predictor parms using default when not found
	predictor, ok := decode_parms.GetInt("Predictor")
	if !ok {
		predictor = 1
	}
	bits_per_component, ok := decode_parms.GetInt("BitsPerComponent")
	if !ok {
		bits_per_component = 8
	}
	colors, ok := decode_parms.GetInt("Colors")
	if !ok {
		colors = 1
	}
	columns, ok := decode_parms.GetInt("Columns")
	if !ok {
		columns = 1
	}

	// make sure bits_per_component value is acceptable
	if bits_per_component <= 0 || bits_per_component > 16 {
		return data
	}

	// determine row widths in bytes
	row_width := (bits_per_component * colors * columns) / 8
	if (bits_per_component * colors * columns) % 8 > 0 {
		row_width++
	}
	if row_width <= 0 {
		return data
	}

	// no predictor applied
	if predictor == 1 {
		return data
	}

	// TIFF predictor
	if predictor == 2 {
		for r := 0; r * row_width < len(data); r ++ {
			row_start := r * row_width * 8
			for c := 1; c < columns; c++ {
				for i := 0; i < colors; i++ {
					pos := row_start + ((c * colors + i) * bits_per_component)
					if pos >= len(data) * 8 {
						return data
					}
					prev_value := GetBits(data, pos - (colors * bits_per_component), bits_per_component)
					value := GetBits(data, pos, bits_per_component)
					SetBits(data, pos, bits_per_component, value + prev_value)
				}
			}
		}
		return data
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
		return decoded_data
	}

	// unknown predictor
	return data
}
