package pdf

import (
	"errors"
)

var ErrEncrypted = errors.New("Pdf is encrypted")
var ErrStartXrefNotFound = errors.New("Missing startxref")
var ErrStreamDictionaryMissing = errors.New("Stream has no stream dictionary")
