package pdf

const (
	KEYWORD_STARTXREF Keyword = iota
	KEYWORD_XREF
	KEYWORD_TRAILER
	KEYWORD_OBJ
	KEYWORD_ENDOBJ
	KEYWORD_STREAM
	KEYWORD_ENDSTREAM
	KEYWORD_R
	KEYWORD_N
	KEYWORD_F
	KEYWORD_NULL
	KEYWORD_TRUE
	KEYWORD_FALSE
	KEYWORD_INVALID
)

type Keyword int

func NewKeyword(keyword string) (Keyword, error) {
	switch keyword {
		case "startxref":
		return KEYWORD_STARTXREF, nil
		case "xref":
		return KEYWORD_XREF, nil
		case "trailer":
		return KEYWORD_TRAILER, nil
		case "obj":
		return KEYWORD_OBJ, nil
		case "endobj":
		return KEYWORD_ENDOBJ, nil
		case "stream":
		return KEYWORD_STREAM, nil
		case "endstream":
		return KEYWORD_ENDSTREAM, nil
		case "R":
		return KEYWORD_R, nil
		case "n":
		return KEYWORD_N, nil
		case "f":
		return KEYWORD_F, nil
		case "null":
		return KEYWORD_NULL, nil
		case "true":
		return KEYWORD_TRUE, nil
		case "false":
		return KEYWORD_FALSE, nil
		default:
		return KEYWORD_INVALID, ErrorKeyword
	}
}

func (keyword Keyword) String() string {
	switch keyword {
		case KEYWORD_STARTXREF:
		return "startxref"
		case KEYWORD_XREF:
		return "xref"
		case KEYWORD_TRAILER:
		return "trailer"
		case KEYWORD_OBJ:
		return "obj"
		case KEYWORD_ENDOBJ:
		return "endobj"
		case KEYWORD_STREAM:
		return "stream"
		case KEYWORD_ENDSTREAM:
		return "endstream"
		case KEYWORD_R:
		return "R"
		case KEYWORD_N:
		return "n"
		case KEYWORD_F:
		return "f"
		case KEYWORD_NULL:
		return "null"
		case KEYWORD_TRUE:
		return "true"
		case KEYWORD_FALSE:
		return "false"
		default:
		return "invalid"
	}
}
