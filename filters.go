package pdf_parser

import (
	"bytes"
	"compress/lzw"
	"compress/zlib"
	"errors"
	"fmt"
	"io"
	"math"
	tiff_lzw "golang.org/x/image/tiff/lzw"
)

func DecodeStream(filter string, data []byte, decode_parms_object Object) ([]byte, error) {
	// convert decode_parms_object to dictionary
	decode_parms, ok := decode_parms_object.(Dictionary)
	if !ok {
		decode_parms = Dictionary{}
	}

	// apply filter
	if filter == "/FlateDecode" {
		return FlateDecode(data, decode_parms)
	} else if filter == "/LZWDecode" {
		return LZWDecode(data, decode_parms)
	} else if filter == "/DCTDecode" {
		// don't bother decoding jpegs into images just dump them as is
		return data, nil
	}

	// return error if filter is not supported
	return data, errors.New(fmt.Sprintf("Unsupported filter: %s", filter))
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
