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
	KEYWORD_TEXT = Keyword("BT")
	KEYWORD_END_TEXT = Keyword("ET")
	KEYWORD_FONT = Keyword("Tf")
	KEYWORD_TEXT_SHOW = Keyword("Tj")
	KEYWORD_TEXT_SHOW_LINE = Keyword("'")
	KEYWORD_TEXT_POSITION = Keyword("TJ")
	KEYWORD_TEXT_POSITION_LINE = Keyword("\"")
	KEYWORD_BEGIN_BF_RANGE = Keyword("beginbfrange")
	KEYWORD_BEGIN_BF_CHAR = Keyword("beginbfchar")
)

type Keyword string

func NewKeyword(keyword string) Keyword {
	return Keyword(keyword)
}

func (keyword Keyword) String() string {
	return string(keyword)
}
