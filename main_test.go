package main

import (
	"github.com/KarmaPenny/pdfparser/pdf"
	"path/filepath"
	"runtime"
	"testing"
)

// LoadTestPdf returns a loaded parser for the provided pdf_name in the test directory
func LoadTestPdf(pdf_name string) (*pdf.Parser, error) {
	// open the pdf
	_, test_path, _, _ := runtime.Caller(0)
	test_dir := filepath.Dir(test_path)
	path := filepath.Join(test_dir, "test", pdf_name)
	parser, err := pdf.Open(path)
	if err != nil {
		return nil, err
	}

	// load the cross reference
	if err := parser.LoadXref(); err != nil {
		// try to repair a bad cross reference
		parser.RepairXref()
	}

	return parser, nil
}

// Make sure the default values for index are used if the index array is not present
func TestXrefStreamMissingIndex(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("xref_stream_missing_index_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 value is correct
	if object.Value.String() != "(Hello)" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

// make sure streams are still read if the stream start marker is only followed by a carriage return
func TestStreamCarriageReturn(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("stream_carriage_return_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 value is correct
	if string(object.Stream) != "Hello" {
		test.Fatalf("Incorrect stream value: %s", string(object.Stream))
	}
}

func TestStrings(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("strings_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 value is correct
	if object.Value.String() != "(newline\nnewline char\nno newline(balance parens allowed) escaped paren ) \\n Hello?)" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestHexStrings(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("hex_strings_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 is "<Hellop>"
	if object.Value.String() != "<Hellop>" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestEmptyHexString(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("empty_hex_string_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 is "<>"
	if object.Value.String() != "<>" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestComments(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("comments_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 is "(%this is not a comment)"
	if object.Value.String() != "(%this is not a comment)" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestReferences(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("references_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 stream is "Hello World"
	if string(object.Stream) != "Hello World" {
		test.Fatalf("Incorrect stream value: %s", string(object.Stream))
	}
}

func TestMultipleFilters(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("multiple_filters_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object 2
	object, err := parser.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 stream is "Hello"
	if string(object.Stream) != "Hello" {
		test.Fatalf("Incorrect stream value: %s", string(object.Stream))
	}
}

func TestMultipleXrefTable(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("multiple_xref_table_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// assert xref length is 10
	if len(parser.Xref) != 10 {
		test.Fatalf("Incorrect number of objects in xref")
	}
}

func TestMultipleXrefStream(test *testing.T) {
	// load the pdf to test
	parser, err := LoadTestPdf("multiple_xref_stream_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// assert xref length is 17
	if len(parser.Xref) != 17 {
		test.Fatalf("Incorrect number of objects in xref")
	}

	// read the third object
	object, err := parser.ReadObject(3)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 3 value is "(Hello)"
	if object.Value.String() != "(Hello)" {
		test.Fatalf("object.Value != (Hello)")
	}
}

func TestASCIIHexDecodeFilter(test *testing.T) {
	// load the pdf to test
	// open pdf for this test
	parser, err := LoadTestPdf("ascii_hex_decode_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the first object
	object, err := parser.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object stream is "Hellop"
	if string(object.Stream) != "Hellop" {
		test.Fatalf("stream != Hellop")
	}
}

func TestASCII85DecodeFilter(test *testing.T) {
	// load the pdf to test
	// open pdf for this test
	parser, err := LoadTestPdf("ascii_85_decode_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the first object
	object, err := parser.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object stream is correct
	if string(object.Stream) != "\x00\x00\x00\x00%!FontType" {
		test.Fatalf("incorrect stream value: %s", string(object.Stream))
	}
}

func TestFlateDecodeFilter(test *testing.T) {
	// load the pdf to test
	// open pdf for this test
	parser, err := LoadTestPdf("flate_decode_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the first object
	_, err = parser.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}
}

func TestLZWDecodeFilter(test *testing.T) {
	// load the pdf to test
	// open pdf for this test
	parser, err := LoadTestPdf("lzw_decode_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the first object
	_, err = parser.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}
}

func TestRunLengthDecodeFilter(test *testing.T) {
	// load the pdf to test
	// open pdf for this test
	parser, err := LoadTestPdf("run_length_decode_test.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the first object
	object, err := parser.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object stream is "Hello"
	if string(object.Stream) != "Hello" {
		test.Fatalf("hello != %s", string(object.Stream))
	}
}
