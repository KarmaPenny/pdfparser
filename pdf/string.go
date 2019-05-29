package pdf

type String string

func (s String) String() string {
	return "(" + string(s) + ")"
}
