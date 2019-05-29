package pdf

import (
	"io"
)

// ReadInt reads width bytes from reader and returns the int64 value
func ReadInt(reader io.Reader, width int) (int, error) {
	v, err := ReadInt64(reader, width)
	return int(v), err
}

// ReadInt reads width bytes from reader and returns the int64 value
func ReadInt64(reader io.Reader, width int) (v int64, err error) {
	data := make([]byte, width)
	bytes_read, _ := reader.Read(data)
	if bytes_read != width {
		return v, NewError("Failed to read int")
	}
	for i := 0; i < len(data); i++ {
		v *= 256
		v += int64(data[i])
	}
	return v, nil
}

// IsHex returns true if the byte is a hex character
func IsHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}
