package pdf

import (
	"fmt"
	"io"
)

type Reference struct {
	parser *Parser
	Number int
	Generation int
}

func NewReference(parser *Parser, number int, generation int) *Reference {
	return &Reference{parser, number, generation}
}

func (reference *Reference) String() string {
	return fmt.Sprintf("%d %d R", reference.Number, reference.Generation)
}

func (reference *Reference) Resolve() Object {
	// save current offset so we can come back
	current_offset := reference.parser.CurrentOffset()

	// resolve the referenced object value
	object := reference.resolve(map[int]interface{}{})

	// revert offset
	reference.parser.Seek(current_offset, io.SeekStart)

	// return the resolved object value
	return object.Value
}

func (reference *Reference) ResolveStream() []byte {
	// save current offset so we can come back
	current_offset := reference.parser.CurrentOffset()

	// resolve the referenced object value
	object := reference.resolve(map[int]interface{}{})

	// revert offset
	reference.parser.Seek(current_offset, io.SeekStart)

	// return the resolved object value
	return object.Stream
}

func (reference *Reference) resolve(resolved_references map[int]interface{}) *IndirectObject {
	// prevent infinite loop
	if _, ok := resolved_references[reference.Number]; ok {
		return NewIndirectObject(reference.Number)
	}
	resolved_references[reference.Number] = nil

	// use parser to get object
	object := reference.parser.GetObject(reference.Number)

	// recursively resolve references
	if ref, ok := object.Value.(*Reference); ok {
		return ref.resolve(resolved_references)
	}

	// return the object
	return object
}
