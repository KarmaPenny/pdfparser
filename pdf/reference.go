package pdf

import (
	"fmt"
	"io"
)

type Reference struct {
	pdf *Parser
	Number int64
	Generation int64
}

func NewReference(pdf *Parser, number int64, generation int64) *Reference {
	return &Reference{pdf, number, generation}
}

func (reference *Reference) String() string {
	return fmt.Sprintf("%d %d R", reference.Number, reference.Generation)
}

func (reference *Reference) Resolve() (Object, error) {
	// save current offset so we can come back
	current_offset, err := reference.pdf.CurrentOffset()
	if err != nil {
		return nil, err
	}

	// resolve the referenced object value
	object := reference.resolve(map[int64]interface{}{})

	// revert offset
	_, err = reference.pdf.Seek(current_offset, io.SeekStart)

	// return the object
	return object, err
}

func (reference *Reference) resolve(resolved_references map[int64]interface{}) Object {
	// prevent infinite loop
	if _, ok := resolved_references[reference.Number]; ok {
		return NewNullObject()
	}
	resolved_references[reference.Number] = nil

	// read the object from the pdf
	object := reference.pdf.ReadObject(reference.Number)

	// recursively resolve references
	if ref, ok := object.Value.(*Reference); ok {
		return ref.resolve(resolved_references)
	}

	// return the object
	return object.Value
}
