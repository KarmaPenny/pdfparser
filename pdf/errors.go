package pdf

import (
	"errors"
)

var EncryptionError = errors.New("missing required encryption info")
var EncryptionPasswordError = errors.New("incorrect password")
var EncryptionUnsupported = errors.New("unsupported encryption")
var EndOfArray = errors.New("end of array")
var EndOfDictionary = errors.New("end of dictionary")
var ReadError = errors.New("read failed")
