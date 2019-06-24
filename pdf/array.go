package pdf

import (
	"strings"
)

type Array []Object

func (a Array) String() string {
	var s strings.Builder
	s.WriteString("[")
	for i, value := range a {
		s.WriteString(value.String())
		if i != len(a) - 1 {
			s.WriteString(" ")
		}
	}
	s.WriteString("]")
	return s.String()
}

func (a Array) GetArray(index int) (Array, bool) {
	if object, ok := a.GetObject(index); ok {
		if array, ok := object.(Array); ok {
			return array, true
		}
	}
	return Array{}, false
}

func (a Array) GetBool(index int) (bool, bool) {
	if object, ok := a.GetObject(index); ok {
		if keyword, ok := object.(Keyword); ok {
			if keyword == KEYWORD_TRUE {
				return true, true
			} else if keyword == KEYWORD_FALSE {
				return false, true
			}
		}
	}
	return false, false
}

func (a Array) GetBytes(index int) ([]byte, bool) {
	s, ok := a.GetString(index)
	return []byte(s), ok
}

func (a Array) GetDictionary(index int) (Dictionary, bool) {
	if object, ok := a.GetObject(index); ok {
		if dictionary, ok := object.(Dictionary); ok {
			return dictionary, true
		}
	}
	return Dictionary{}, false
}

func (a Array) GetInt(index int) (int, bool) {
	number, ok := a.GetNumber(index)
	return int(number), ok
}

func (a Array) GetInt64(index int) (int64, bool) {
	number, ok := a.GetNumber(index)
	return int64(number), ok
}

func (a Array) GetName(index int) (string, bool) {
	if object, ok := a.GetObject(index); ok {
		if name, ok := object.(Name); ok {
			return string(name), true
		}
	}
	return "", false
}

func (a Array) GetNumber(index int) (Number, bool) {
	if object, ok := a.GetObject(index); ok {
		if number, ok := object.(Number); ok {
			return number, true
		}
	}
	return Number(0), false
}

func (a Array) GetObject(index int) (Object, bool) {
	if index >= 0 && index < len(a) {
		object := a[index]
		if reference, ok := object.(*Reference); ok {
			return reference.Resolve(), true
		}
		return object, true
	}
	return KEYWORD_NULL, false
}

func (a Array) GetStream(index int) ([]byte, bool) {
	if index >= 0 && index < len(a) {
		if reference, ok := a[index].(*Reference); ok {
			return reference.ResolveStream(), true
		}
	}
	return []byte{}, false
}

func (a Array) GetString(index int) (string, bool) {
	if object, ok := a.GetObject(index); ok {
		if s, ok := object.(String); ok {
			return string(s), true
		}
	}
	return "", false
}
