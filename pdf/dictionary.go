package pdf

import (
	"strings"
)

type Dictionary map[string]Object

func (d Dictionary) String() string {
	var s strings.Builder
	s.WriteString("<<")
	for key, value := range d {
		s.WriteString("/")
		s.WriteString(key)
		s.WriteString(" ")
		s.WriteString(value.String())
	}
	s.WriteString(">>")
	return s.String()
}

func (d Dictionary) Contains(key string) bool {
	_, ok := d[key]
	return ok
}

func (d Dictionary) GetArray(key string) (Array, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return Array{}, err
	}
	if array, ok := object.(Array); ok {
		return array, nil
	}
	return Array{}, NewError("Expected array")
}

func (d Dictionary) GetBool(key string) (bool, error) {
	object, err := d.GetObject(key)
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

func (d Dictionary) GetDictionary(key string) (Dictionary, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return Dictionary{}, err
	}
	if dictionary, ok := object.(Dictionary); ok {
		return dictionary, nil
	}
	return Dictionary{}, NewError("Expected dictionary")
}

func (d Dictionary) GetInt(key string) (int, error) {
	number, err := d.GetNumber(key)
	return int(number), err
}

func (d Dictionary) GetInt64(key string) (int64, error) {
	number, err := d.GetNumber(key)
	return int64(number), err
}

func (d Dictionary) GetName(key string) (string, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return "", err
	}
	if name, ok := object.(Name); ok {
		return string(name), nil
	}
	return "", NewError("Expected name")
}

func (d Dictionary) GetNumber(key string) (Number, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return Number(0), err
	}
	if number, ok := object.(Number); ok {
		return number, nil
	}
	return Number(0), NewError("Expected number")
}

func (d Dictionary) GetObject(key string) (Object, error) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return reference.Resolve(), nil
		}
		return object, nil
	}
	return KEYWORD_NULL, NewError("Dictionary does not contain key: %s", key)
}

func (d Dictionary) GetString(key string) (string, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return "", err
	}
	if s, ok := object.(String); ok {
		return string(s), nil
	}
	return "", NewError("Expected string")
}
