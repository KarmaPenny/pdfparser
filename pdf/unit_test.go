package pdf

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func openTestPdf(pdf_name string) (*os.File, error) {
	_, test_path, _, _ := runtime.Caller(0)
	test_dir := filepath.Dir(test_path)
	return os.Open(filepath.Join(test_dir, "test", pdf_name))
}

func TestComments(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("comments.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if object.Value.String() != "(%this is not a comment)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestEmptyArray(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("empty_array.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if object.Value.String() != "[]" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestEmptyDictionary(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("empty_dictionary.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if object.Value.String() != "<<>>" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestEncrypted(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("encrypted.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// test string decryption
	object := parser.GetObject(12)
	d12, ok := object.Value.(Dictionary)
	if !ok {
		test.Fatalf("expected dictionary")
	}
	if lang, _ := d12.GetString("Lang"); lang != "en-US" {
		test.Fatalf("incorrect value %s", lang)
	}

	// test stream decryption
	object = parser.GetObject(8)
	if string(object.Stream[:8]) != "/CIDInit" {
		test.Fatalf("incorrect value %s", string(object.Stream[:8]))
	}
}

func TestFilterASCII85Decode(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_ascii_85_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "\x00\x00\x00\x00%!FontType" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterASCIIHexDecode(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_ascii_hex_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "Hellop" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterFlateDecode(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_flate_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "hello world\nhello world\nhello world\nhello world\nhello world\n" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterLZWDecode(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_lzw_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "hello world!" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterLZWTiffDecode(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_lzw_tiff_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "hello world!" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterMultiple(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_multiple.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "hello world\nhello world\nhello world\nhello world\nhello world\n" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestFilterRunLengthDecode(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("filter_run_length_decode.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "Hello" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestMalformedDictionaryKey(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("malformed_dictionary_key.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read object
	object := parser.GetObject(1)

	// assert value is correct
	if d, ok := object.Value.(Dictionary); ok {
		if hidden, _ := d.GetString("HiddenObject"); hidden != "Hello World" {
			test.Fatalf("incorrect value %s", hidden)
		}
	}
}

func TestNames(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("names.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read object
	object := parser.GetObject(1)

	// assert value is correct
	if object.Value.String() != "/Hello /World!\x00qz" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestReference(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("reference.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read object
	object := parser.GetObject(1)

	// make sure object is a reference
	reference, ok := object.Value.(*Reference)
	if !ok {
		test.Fatal("object is not a reference")
	}

	// resolve the reference
	resolved_object := reference.Resolve()

	// assert value is correct
	if resolved_object.String() != "(Hello World!)" {
		test.Fatalf("incorrect value %s", resolved_object.String())
	}
}

func TestReferenceLoop(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("reference_loop.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read object
		object := parser.GetObject(1)

		// make sure object is a reference
		reference, ok := object.Value.(*Reference)
		if !ok {
			test.Fatal("object is not a reference")
		}

		// resolve the reference
		resolved_object := reference.Resolve()

		// assert value is correct
		if resolved_object.String() != "null" {
			test.Fatalf("incorrect value %s", resolved_object.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestReferenceNull(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("reference_null.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read object
	object := parser.GetObject(1)

	// make sure object is a reference
	reference, ok := object.Value.(*Reference)
	if !ok {
		test.Fatal("object is not a reference")
	}

	// resolve the reference
	resolved_object := reference.Resolve()

	// assert value is correct
	if resolved_object.String() != "null" {
		test.Fatalf("incorrect value %s", resolved_object.String())
	}
}

func TestCarriageReturn(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("carriage_return.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// assert value is correct
	if string(object.Stream) != "Hello" {
		test.Fatalf("incorrect value %s", string(object.Stream))
	}
}

func TestStrings(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("strings.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// read the object
	object := parser.GetObject(1)

	// make sure object is an array
	array, ok := object.Value.(Array)
	if !ok {
		test.Fatal("object is not an Array")
	}

	// assert values are correct
	if array[0].String() != "(newline\nnewline char\nno newline(balance parens allowed) escaped paren ) \\n Hello??7)" {
		test.Fatalf("incorrect value %s", array[0].String())
	}
	if array[1].String() != "(Hellop)" {
		test.Fatalf("incorrect value %s", array[1].String())
	}
	if array[2].String() != "()" {
		test.Fatalf("incorrect value %s", array[2].String())
	}
	if array[3].String() != "()" {
		test.Fatalf("incorrect value %s", array[2].String())
	}
}

func TestUnclosedArray(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_array.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "[]" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedComment(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_comment.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "null" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedDictionary(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_dictionary.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		parser.GetObject(1)
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}
func TestUnclosedDictionaryKey(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_dictionary_key.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		parser.GetObject(1)
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedHexString(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_hex_string.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "()" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedName(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_name.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "/" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedNameEscape1(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_name_escape_1.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "/\x00" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedNameEscape2(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_name_escape_2.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "/0" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedString(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_string.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "()" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedStringEscape(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_string_escape.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "(\\)" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedStringOctal1(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_string_octal_1.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "(\x01)" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestUnclosedStringOctal2(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("unclosed_string_octal_2.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// read the object
		object := parser.GetObject(1)

		// make sure object value is correct
		if object.Value.String() != "(\n)" {
			test.Fatalf("incorrect value %s", object.Value.String())
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestXrefLoop(test *testing.T) {
	// create a done signal channel
	done := make(chan bool, 1)

	// run test in background
	go func() {
		// send done signal when returning
		defer func() {done <- true}()

		// open the pdf
		f, err := openTestPdf("xref_loop.pdf")
		if err != nil {
			test.Fatal(err)
		}
		defer f.Close()

		// load the pdf
		parser := NewParser(f)
		err = parser.Load("")
		if err != nil {
			test.Fatal(err)
		}

		// assert xref length is correct
		if len(parser.Xref) != 10 {
			test.Fatal("xref length != 10")
		}
	}()

	// wait until done or timed out
	select {
		case <-done:
		case <-time.After(time.Second):
			test.Fatal("timed out")
	}
}

func TestXrefRepair(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("xref_repair.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// assert xref length is correct
	if len(parser.Xref) != 9 {
		test.Fatalf("%d != 9", len(parser.Xref))
	}

	// read the object
	object := parser.GetObject(9)

	// assert values are correct
	if object.Value.String() != "(Hello world)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestXrefStreamChain(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("xref_stream_chain.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// assert xref length is correct
	if len(parser.Xref) != 11 {
		test.Fatalf("%d != 11", len(parser.Xref))
	}

	// read the object
	object := parser.GetObject(10)

	// assert values are correct
	if object.Value.String() != "(Hello World!)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestXrefStreamIndexDefault(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("xref_stream_index_default.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// assert xref length is correct
	if len(parser.Xref) != 10 {
		test.Fatalf("%d != 10", len(parser.Xref))
	}

	// read the object
	object := parser.GetObject(9)

	// assert values are correct
	if object.Value.String() != "(Hello World!)" {
		test.Fatalf("incorrect value %s", object.Value.String())
	}
}

func TestXrefTableChain(test *testing.T) {
	// open the pdf
	f, err := openTestPdf("xref_table_chain.pdf")
	if err != nil {
		test.Fatal(err)
	}
	defer f.Close()

	// load the pdf
	parser := NewParser(f)
	err = parser.Load("")
	if err != nil {
		test.Fatal(err)
	}

	// assert xref length is correct
	if len(parser.Xref) != 10 {
		test.Fatal("xref length != 10")
	}
}
