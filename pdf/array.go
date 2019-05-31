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

func (a Array) GetArray(index int) (Array, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return Array{}, err
	}
	if array, ok := object.(Array); ok {
		return array, nil
	}
	return Array{}, NewError("Expected array")
}

func (a Array) GetBool(index int) (bool, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return false, err
	}
	if keyword, ok := object.(Keyword); ok {
		if keyword == KEYWORD_TRUE {
			return true, nil
		} else if keyword == KEYWORD_FALSE {
			return false, nil
		}
	}
	return false, NewError("Expected bool")
}

func (a Array) GetBytes(index int) ([]byte, error) {
	s, err := a.GetString(index)
	return []byte(s), err
}

func (a Array) GetDictionary(index int) (Dictionary, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return Dictionary{}, err
	}
	if dictionary, ok := object.(Dictionary); ok {
		return dictionary, nil
	}
	return Dictionary{}, NewError("Expected dictionary")
}

func (a Array) GetInt(index int) (int, error) {
	number, err := a.GetNumber(index)
	return int(number), err
}

func (a Array) GetInt64(index int) (int64, error) {
	number, err := a.GetNumber(index)
	return int64(number), err
}

func (a Array) GetName(index int) (string, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return "", err
	}
	if name, ok := object.(Name); ok {
		return string(name), nil
	}
	return "", NewError("Expected name")
}

func (a Array) GetNumber(index int) (Number, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return Number(0), err
	}
	if number, ok := object.(Number); ok {
		return number, nil
	}
	return Number(0), NewError("Expected number")
}

func (a Array) GetObject(index int) (Object, error) {
	if index < 0 || index >= len(a) {
		return KEYWORD_NULL, NewError("index out of bounds")
	}
	object := a[index]
	if reference, ok := object.(*Reference); ok {
		return reference.Resolve(), nil
	}
	return object, nil
}

func (a Array) GetString(index int) (string, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return "", err
	}
	if s, ok := object.(String); ok {
		return string(s), nil
	}
	return "", NewError("Expected string")
}
