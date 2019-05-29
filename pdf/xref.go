package pdf

const (
	XrefTypeFreeObject = iota
	XrefTypeIndirectObject
	XrefTypeCompressedObject
)

type XrefEntry struct {
	Offset int64
	Generation int
	Type int
}

func NewXrefEntry(offset int64, generation int, type_value int) *XrefEntry {
	return &XrefEntry{offset, generation, type_value}
}
