package pdf

import (
	"io"
)

// ReadInt reads width bytes from reader and returns the int64 value
func ReadInt(reader io.Reader, width int64) (v int64, err error) {
	data := make([]byte, width)
	bytes_read, err := reader.Read(data)
	if int64(bytes_read) != width && err != nil {
		return v, WrapError(err, "Failed to read int64")
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
