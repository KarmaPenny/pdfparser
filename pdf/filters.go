package pdf

import (
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"encoding/ascii85"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	tiff_lzw "golang.org/x/image/tiff/lzw"
)

func DecodeStream(filter string, data []byte, decode_parms_object Object) ([]byte, error) {
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

	// don't bother decoding jpegs into images just dump them as is
	if filter == "/DCTDecode" {
		return data, nil
	}

	// return error if filter is not supported
	return data, errors.New(fmt.Sprintf("Unsupported filter: %s", filter))
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
			return decoded_data, errors.New("illegal character in ASCIIHexDecode stream")
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
				return decoded_data, errors.New("illegal character in ASCIIHexDecode stream")
			}

			// set second byte
			i++
			b2 = data[i]
			break
		}

		// add decoded byte to decoded data
		val, err := strconv.ParseUint(string([]byte{b1, b2}), 16, 8)
		if err != nil {
			return decoded_data, errors.New(fmt.Sprintf("Hex decode failed: %s", err))
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
	decoder := ascii85.NewDecoder(bytes.NewReader(data))
	var decoded_data bytes.Buffer
	bytes_read, err := decoded_data.ReadFrom(decoder)
	if bytes_read == 0 && err != nil {
		return data, errors.New(fmt.Sprintf("failed to ascii85 decode stream: %s", err))
	}
	return decoded_data.Bytes(), nil
}

func RunLengthDecode(data []byte) ([]byte, error) {
	var decoded_data bytes.Buffer
	for i := 0; i < len(data); {
		// get length byte
		length := int(data[i])

		// EOD
		if length == 128 {
			break
		}

		// literal copy
		if length < 128 {
			// length is value of byte plus one
			length++

			// increment index
			i++

			// make sure slice is within data bounds
			if i + length > len(data) {
				return decoded_data.Bytes(), fmt.Errorf("Run Length error: Not enough data to copy")
			}

			// copy the next length bytes to decoded_data
			decoded_data.Write(data[i:i + length])

			// increment index by length and continue
			i += length
			continue
		}

		// duplicate next byte
		if length > 128 {
			// increment index
			i++

			// make sure byte is within data bounds
			if i >= len(data) {
				return decoded_data.Bytes(), errors.New("Run Length error: byte out of bounds")
			}

			times := 257 - length

			// copy times times
			for n := 0; n < times; n++ {
				decoded_data.WriteByte(data[i])
			}

			// increment index
			i++
		}
	}
	return decoded_data.Bytes(), nil
}

func FlateDecode(data []byte, decode_parms Dictionary) ([]byte, error) {
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
		return data, errors.New(fmt.Sprintf("failed to lzw decode stream: %s", err))
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
