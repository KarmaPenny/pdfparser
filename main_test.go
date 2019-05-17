package main

import (
	"github.com/KarmaPenny/pdfparser/pdf"
	"path/filepath"
	"runtime"
	"testing"
)

// GetPath returns the path to a pdf in the test pdfs dir
func GetPath(pdf_name string) string {
	_, test_path, _, _ := runtime.Caller(0)
	test_dir := filepath.Dir(test_path)
	return filepath.Join(test_dir, "test", pdf_name)
}

func TestStrings(test *testing.T) {
	reader, err := pdf.Open(GetPath("strings_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read the object 2
	object, err := reader.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 value is correct
	if object.Value.String() != "(newline\nnewline char\nno newline(balance parens allowed) escaped paren ) \\n Hello?)" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestHexStrings(test *testing.T) {
	reader, err := pdf.Open(GetPath("hex_strings_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read object 2
	object, err := reader.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 is "<Hellop>"
	if object.Value.String() != "<Hellop>" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestComments(test *testing.T) {
	reader, err := pdf.Open(GetPath("comments_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read object 2
	object, err := reader.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 is "(%this is not a comment)"
	if object.Value.String() != "(%this is not a comment)" {
		test.Fatalf("Incorrect string value: %s", object.Value.String())
	}
}

func TestReferences(test *testing.T) {
	reader, err := pdf.Open(GetPath("references_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read object 2
	object, err := reader.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 stream is "Hello World"
	if string(object.Stream) != "Hello World" {
		test.Fatalf("Incorrect stream value: %s", string(object.Stream))
	}
}

func TestMultipleFilters(test *testing.T) {
	reader, err := pdf.Open(GetPath("multiple_filters_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read object 2
	object, err := reader.ReadObject(2)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 2 stream is "Hello"
	if string(object.Stream) != "Hello" {
		test.Fatalf("Incorrect stream value: %s", string(object.Stream))
	}
}

func TestMultipleXrefTable(test *testing.T) {
	reader, err := pdf.Open(GetPath("multiple_xref_table_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// assert xref length is 10
	if len(reader.Xref) != 10 {
		test.Fatalf("Incorrect number of objects in xref")
	}
}

func TestMultipleXrefStream(test *testing.T) {
	reader, err := pdf.Open(GetPath("multiple_xref_stream_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// assert xref length is 17
	if len(reader.Xref) != 17 {
		test.Fatalf("Incorrect number of objects in xref")
	}

	// read the third object
	object, err := reader.ReadObject(3)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object 3 value is "(Hello)"
	if object.Value.String() != "(Hello)" {
		test.Fatalf("object.Value != (Hello)")
	}
}

// TestEncryptedPdf asserts that encrypted pdfs return pdf.ErrEncrypted on opening
func TestEncryptedPdf(test *testing.T) {
	reader, err := pdf.Open(GetPath("encrypted_test.pdf"))
	if err != pdf.ErrEncrypted {
		if err == nil {
			reader.Close()
		}
		test.Fatalf("Failed to detect encryption: %s", err)
	}
}

func TestASCIIHexDecodeFilter(test *testing.T) {
	// open pdf for this test
	reader, err := pdf.Open(GetPath("ascii_hex_decode_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read the first object
	object, err := reader.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object stream is "Hellop"
	if string(object.Stream) != "Hellop" {
		test.Fatalf("stream != Hellop")
	}
}

func TestASCII85DecodeFilter(test *testing.T) {
	// open pdf for this test
	reader, err := pdf.Open(GetPath("ascii_85_decode_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read the first object
	_, err = reader.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}
}

func TestFlateDecodeFilter(test *testing.T) {
	// open pdf for this test
	reader, err := pdf.Open(GetPath("flate_decode_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read the first object
	_, err = reader.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}
}

func TestLZWDecodeFilter(test *testing.T) {
	// open pdf for this test
	reader, err := pdf.Open(GetPath("lzw_decode_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read the first object
	_, err = reader.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}
}

func TestRunLengthDecodeFilter(test *testing.T) {
	// open pdf for this test
	reader, err := pdf.Open(GetPath("run_length_decode_test.pdf"))
	if err != nil {
		test.Fatalf("Failed to open pdf: %s", err)
	}
	defer reader.Close()

	// read the first object
	object, err := reader.ReadObject(1)
	if err != nil {
		test.Fatalf("Failed to read object: %s", err)
	}

	// assert object stream is "Hello"
	if string(object.Stream) != "Hello" {
		test.Fatalf("hello != %s", string(object.Stream))
	}
}
