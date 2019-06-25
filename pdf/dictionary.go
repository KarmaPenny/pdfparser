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

func (d Dictionary) GetArray(key string) (Array, bool) {
	if object, ok := d.GetObject(key); ok {
		if array, ok := object.(Array); ok {
			return array, true
		}
	}
	return Array{}, false
}

func (d Dictionary) GetBool(key string) (bool, bool) {
	if object, ok := d.GetObject(key); ok {
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

func (d Dictionary) GetBytes(key string) ([]byte, bool) {
	s, ok := d.GetString(key)
	return []byte(s), ok
}

func (d Dictionary) GetDictionary(key string) (Dictionary, bool) {
	if object, ok := d.GetObject(key); ok {
		if dictionary, ok := object.(Dictionary); ok {
			return dictionary, true
		}
	}
	return Dictionary{}, false
}

func (d Dictionary) GetInt(key string) (int, bool) {
	number, ok := d.GetNumber(key)
	return int(number), ok
}

func (d Dictionary) GetInt64(key string) (int64, bool) {
	number, ok := d.GetNumber(key)
	return int64(number), ok
}

func (d Dictionary) GetName(key string) (string, bool) {
	if object, ok := d.GetObject(key); ok {
		if name, ok := object.(Name); ok {
			return string(name), true
		}
	}
	return "", false
}

func (d Dictionary) GetNameTreeMap(key string) Array {
	if root, ok := d.GetDictionary(key); ok {
		return root.getNameTreeMap(map[int]interface{}{})
	}
	return Array{}
}

func (d Dictionary) getNameTreeMap(resolved_kids map[int]interface{}) Array {
	nameTreeMap := Array{}
	if names, ok := d.GetArray("Names"); ok {
		// append names to name tree map
		nameTreeMap = append(nameTreeMap, names...)
	}
	if kids, ok := d.GetArray("Kids"); ok {
		for i := range kids {
			// prevent infinite resolve reference loop
			if r, ok := kids[i].(*Reference); ok {
				if _, resolved := resolved_kids[r.Number]; resolved {
					continue
				}
				resolved_kids[r.Number] = nil
			}

			// get child dictionary
			kid, _ := kids.GetDictionary(i)
			nameTreeMap = append(nameTreeMap, kid.getNameTreeMap(resolved_kids)...)
		}
	}
	return nameTreeMap
}

func (d Dictionary) GetNumber(key string) (Number, bool) {
	if object, ok := d.GetObject(key); ok {
		if number, ok := object.(Number); ok {
			return number, true
		}
	}
	return Number(0), false
}

func (d Dictionary) GetObject(key string) (Object, bool) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return reference.Resolve(), true
		}
		return object, true
	}
	return KEYWORD_NULL, false
}

func (d Dictionary) GetPageTree(key string) ([]Dictionary, bool) {
	if pageTree, ok := d.GetDictionary(key); ok {
		return pageTree.ResolveKids(map[int]interface{}{}), true
	}
	return []Dictionary{}, false
}

func (d Dictionary) ResolveKids(resolved_kids map[int]interface{}) []Dictionary {
	kids_list := []Dictionary{d}
	if kids, ok := d.GetArray("Kids"); ok {
		for i := range kids {
			if r, ok := kids[i].(*Reference); ok {
				// ignore circular references
				if _, resolved := resolved_kids[r.Number]; resolved {
					continue
				}
				resolved_kids[r.Number] = nil

				// resolve
				object := r.Resolve()
				if kid_d, ok := object.(Dictionary); ok {
					// recursively resolve kids
					kids_list = append(kids_list, kid_d.ResolveKids(resolved_kids)...)
				}
			}
		}
	}
	return kids_list
}

func (d Dictionary) GetReference(key string) (*Reference, bool) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return reference, true
		}
	}
	return nil, false
}

func (d Dictionary) GetStream(key string) ([]byte, bool) {
	if object, ok := d[key]; ok {
		if reference, ok := object.(*Reference); ok {
			return reference.ResolveStream(), true
		}
	}
	return []byte{}, false
}

func (d Dictionary) GetString(key string) (string, bool) {
	if object, ok := d.GetObject(key); ok {
		if s, ok := object.(String); ok {
			return string(s), true
		}
	}
	return "", false
}
