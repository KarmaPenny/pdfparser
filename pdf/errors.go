package pdf

import (
	"errors"
)


// errors
var EncryptionError = errors.New("missing required encryption info")
var EncryptionPasswordError = errors.New("incorrect password")
var EncryptionUnsupported = errors.New("unsupported encryption")
var EndOfArray = errors.New("end of array")
var EndOfDictionary = errors.New("end of dictionary")
var ReadError = errors.New("read failed")

// format errors and abnormalities
var InvalidDictionaryKeyType = "invalid dictionary key type"
var InvalidHexStringChar = "invalid hex string character"
var InvalidNameEscapeChar = "invalid name escape character"
var InvalidOctal = "invalid octal in string"
var MissingDictionaryValue = "missing dictionary value"
var UnclosedArray = "unclosed array"
var UnclosedDictionary = "unclosed dictionary"
var UnclosedHexString = "unclosed hex string"
var UnclosedStream = "unclosed stream"
var UnclosedString = "unclosed string"
var UnclosedStringEscape = "unclosed escape in string"
var UnclosedStringOctal = "unclosed octal in string"
var UnnecessaryEscapeName = "unnecessary espace sequence in name"
var UnnecessaryEscapeString = "unnecessary espace sequence in string"
