package pdf

import (
	"io"
)

// ReadInt reads width bytes from reader and returns the int64 value
func ReadInt(reader io.Reader, width int) (int, bool) {
	v, ok := ReadInt64(reader, width)
	return int(v), ok
}

// ReadInt reads width bytes from reader and returns the int64 value
func ReadInt64(reader io.Reader, width int) (int64, bool) {
	data := make([]byte, width)
	bytes_read, _ := reader.Read(data)
	if bytes_read != width {
		return 0, false
	}
	v := int64(0)
	for i := 0; i < len(data); i++ {
		v *= 256
		v += int64(data[i])
	}
	return v, true
}

// IsHex returns true if the byte is a hex character
func IsHex(b byte) bool {
	return (b >= '0' && b <= '9') || (b >= 'a' && b <= 'f') || (b >= 'A' && b <= 'F')
}

// GetBits gets n bits from d starting at p
func GetBits(d []byte, p int, n int) uint32 {
	var v uint32 = 0
	for i := 0; i < 4; i++ {
		v *= 256
		if p / 8 + i < len(d) {
			v += uint32(d[p / 8 + i])
		}
	}
	v <<= uint(p % 8)
	v >>= uint(32 - n)
	return v
}

// SetBits sets n bits in d starting at p to the last n bits of v
func SetBits(d []byte, p int, n int, v uint32) {
	dv := GetBits(d, (p / 8) * 8, 32)
	s := uint(32 - n - (p % 8))
	max := uint32(1 << uint(n))
	m := uint32(1 << uint(n)) - 1
	dv += ((v % max) << s) - (dv & (m << s))
	for i := 0; i < 4 && p / 8 + i < len(d); i++ {
		d[p / 8 + i] = byte(dv >> (32 - ((uint(i) + 1) * 8)))
	}
}

func BytesToInt(b []byte) int {
	val := 0
	for i := range b {
		val *= 256
		val += int(b[i])
	}
	return val
}
