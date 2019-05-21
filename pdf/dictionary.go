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
		return 0, WrapError(err, "Dictionary value at %s is not int32: %s", key, object.String())
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
		return 0, WrapError(err, "Dictionary value at %s is not int64: %s", key, object.String())
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
	return nil, WrapError(err, "Dictionary value at %s is not Array: %s", key, object.String())
}

func (d Dictionary) GetDictionary(key string) (Dictionary, error) {
	object, err := d.GetObject(key)
	if err != nil {
		return nil, err
	}
	if dictionary, ok := object.(Dictionary); ok {
		return dictionary, nil
	}
	return nil, WrapError(err, "Dictionary value at %s is not Dictionary: %s", key, object.String())
}

// GetObjectFromDictionary gets an object from a dictionary, resolving references if needed
func (d Dictionary) GetObject(key string) (Object, error) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return reference.Resolve()
		}
		return object, nil
	}
	return nil, NewError("Dictionary does not contain key: %s", key)
}
