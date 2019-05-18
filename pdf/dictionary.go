package pdf

import (
	"strconv"
	"strings"
)

type Dictionary map[string]Object

func (d Dictionary) String() string {
	var s strings.Builder
	s.WriteString("<<")
	for key, value := range d {
		s.WriteString(key)
		s.WriteString(" ")
		s.WriteString(value.String())
	}
	s.WriteString(">>")
	return s.String()
}

func (d Dictionary) GetInt(key string) (int, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(object.String(), 10, 32)
	if err != nil {
		return 0, NewError(err)
	}
	return int(i), nil
}

func (d Dictionary) GetInt64(key string) (int64, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(object.String(), 10, 64)
	if err != nil {
		return 0, NewError(err)
	}
	return i, nil
}

func (d Dictionary) GetArray(key string) (Array, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return nil, err
	}
	if array, ok := object.(Array); ok {
		return array, nil
	}
	return nil, NewErrorf("Expected Array")
}

// GetObjectFromDictionary gets an object from a dictionary, resolving references if needed
func (d Dictionary) GetObject(key string) (Object, error) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return reference.Resolve()
		}
		return object, nil
	}
	return nil, NewErrorf("Missing key: %s", key)
}
