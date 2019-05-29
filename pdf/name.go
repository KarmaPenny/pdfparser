package pdf

type Name string

func (name Name) String() string {
	return "/" + string(name)
}
