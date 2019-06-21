package pdf

const (
	KEYWORD_XREF = Keyword("xref")
	KEYWORD_TRAILER = Keyword("trailer")
	KEYWORD_OBJ = Keyword("obj")
	KEYWORD_STREAM = Keyword("stream")
	KEYWORD_R = Keyword("R")
	KEYWORD_N = Keyword("n")
	KEYWORD_NULL = Keyword("null")
	KEYWORD_TRUE = Keyword("true")
	KEYWORD_FALSE = Keyword("false")
)

type Keyword string

func NewKeyword(keyword string) Keyword {
	return Keyword(keyword)
}

func (keyword Keyword) String() string {
	return string(keyword)
}
