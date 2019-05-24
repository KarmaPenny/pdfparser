package main

import (
	"github.com/KarmaPenny/pdfparser/pdf"
	"path/filepath"
	"runtime"
	"testing"
)

// OpenTestPdf returns a loaded parser for the provided pdf_name in the test directory
func OpenTestPdf(pdf_name string) (*pdf.Parser, error) {
	_, test_path, _, _ := runtime.Caller(0)
	test_dir := filepath.Dir(test_path)
	path := filepath.Join(test_dir, "test", pdf_name)
	return pdf.Open(path)
}

func TestComments(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("comments.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if object.Value.String() != "(%this is not a comment)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestFilterASCII85Decode(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_ascii_85_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "\x00\x00\x00\x00%!FontType" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterASCIIHexDecode(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_ascii_hex_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "Hellop" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterFlateDecode(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_flate_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "hello world\nhello world\nhello world\nhello world\nhello world\n" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterLZWDecode(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_lzw_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "hello world!" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterLZWTiffDecode(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_lzw_tiff_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "hello world!" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterMultiple(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_multiple.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "hello world\nhello world\nhello world\nhello world\nhello world\n" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterRunLengthDecode(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("filter_run_length_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "Hello" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestNames(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("names.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if object.Value.String() != "/Hello /World!" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestReferences(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("references.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object
	object, err := parser.ReadObject(1)
	if err != nil {
		test.Fatal(err)
	}

	// make sure object is a reference
	reference, ok := object.Value.(*pdf.Reference)
	if !ok {
		test.Fatal("object is not a reference")
	}

	// resolve the reference
	resolved_object, err := reference.Resolve()
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if resolved_object.String() != "(Hello World!)" {
		test.Fatalf("incorrect value %s", resolved_object.String())
	}
}

func TestReferencesInfiniteLoop(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("references.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read object
	object, err := parser.ReadObject(3)
	if err != nil {
		test.Fatal(err)
	}

	// make sure object is a reference
	reference, ok := object.Value.(*pdf.Reference)
	if !ok {
		test.Fatal("object is not a reference")
	}

	// resolve the reference
	resolved_object, err := reference.Resolve()
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if resolved_object.String() != "null" {
		test.Fatalf("incorrect value %s", resolved_object.String())
	}
}

func TestStreamCarriageReturn(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("stream_carriage_return.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert value is correct
	if string(object.Stream) != "Hello" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestStrings(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("strings.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// make sure object is an array
	array, ok := object.Value.(pdf.Array)
	if !ok {
		test.Fatal("object is not an Array")
	}

	// assert values are correct
	if array[0].String() != "(newline\nnewline char\nno newline(balance parens allowed) escaped paren ) \\n Hello?)" {
		test.Fatalf("incorrect value %s", array[0].String())
	}
	if array[1].String() != "<Hellop>" {
		test.Fatalf("incorrect value %s", array[1].String())
	}
	if array[2].String() != "<>" {
		test.Fatalf("incorrect value %s", array[2].String())
	}
}

func TestXrefRepair(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("xref_repair.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// assert xref length is correct
	if len(parser.Xref) != 9 {
		test.Fatalf("%d != 9", len(parser.Xref))
	}

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert values are correct
	if object.Value.String() != "(Hello world)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestXrefStreamChain(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("xref_stream_chain.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// assert xref length is correct
	if len(parser.Xref) != 11 {
		test.Fatalf("%d != 11", len(parser.Xref))
	}

	// read the object
	object, err := parser.ReadObject(10)
	if err != nil {
		test.Fatal(err)
	}

	// assert values are correct
	if object.Value.String() != "(Hello World!)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestXrefStreamIndexDefault(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("xref_stream_index_default.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// assert xref length is correct
	if len(parser.Xref) != 10 {
		test.Fatalf("%d != 10", len(parser.Xref))
	}

	// read the object
	object, err := parser.ReadObject(9)
	if err != nil {
		test.Fatal(err)
	}

	// assert values are correct
	if object.Value.String() != "(Hello World!)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestXrefTableChain(test *testing.T) {
	// open the pdf
	parser, err := OpenTestPdf("xref_table_chain.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer parser.Close()

	// assert xref length is correct
	if len(parser.Xref) != 10 {
		test.Fatal("xref length != 10")
	}
}
