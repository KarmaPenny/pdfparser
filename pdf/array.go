package pdf

import (
	"strconv"
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

func (a Array) GetInt(index int) (int, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(object.String(), 10, 32)
	if err != nil {
		return 0, WrapError(err, "Array element at %d is not an int32: %s", index, object.String())
	}
	return int(i), nil
}

func (a Array) GetInt64(index int) (int64, error) {
	object, err := a.GetObject(index)
	if err != nil {
		return 0, err
	}
	i, err := strconv.ParseInt(object.String(), 10, 64)
	if err != nil {
		return 0, WrapError(err, "Array element at %d is not an int64: %s", index, object.String())
	}
	return i, nil
}

// GetObject gets an object from an array, resolving references if needed
func (a Array) GetObject(index int) (Object, error) {
	if index < len(a) || index < 0 {
		object := a[index]
		if reference, ok := object.(*Reference); ok {
			return reference.Resolve()
		}
		return object, nil
	}
	return nil, NewError("Array index [%d] out of bounds: %d", index, len(a))
}
