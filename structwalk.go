package structwalk

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

// FieldValue returns a value of a field at path in deeply nested structs, will traverse both maps and structs.
func FieldValue(path string, in interface{}) (v interface{}, found bool) {
	var cur reflect.Value

	cur, found = findValue(path, in)
	if !found {
		return
	}

	v, found = cur.Interface(), true

	return
}

// SetFieldValue sets a value of a field at path in deeply nested structs, will traverse both maps and structs.
func SetFieldValue(path string, v interface{}, in interface{}) {
	var cur reflect.Value

	cur, found := findValue(path, in)
	if !found {
		return
	}

	cur.Set(reflect.ValueOf(v))
}

func findValue(path string, in interface{}) (v reflect.Value, found bool) {
	defer func() {
		if x := recover(); x != nil {
			return
		}
	}()

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return reflect.Value{}, false
	}

	cur := reflect.ValueOf(in)
	for i, part := range parts {
		part := strings.ToLower(part)

		for {
			if cur.Kind() == reflect.Ptr || cur.Kind() == reflect.Interface {
				cur = cur.Elem()
				continue
			}
			break
		}

		if cur.Kind() == reflect.Struct {
			cur = cur.FieldByNameFunc(func(name string) bool {
				return strings.ToLower(name) == part
			})
		} else if cur.Kind() == reflect.Map {
			keys := cur.MapKeys()
			var keyFound bool
			for _, k := range keys {
				if strings.ToLower(k.String()) == part {
					cur = cur.MapIndex(k)
					keyFound = true
					break
				}
			}
			if !keyFound {
				return reflect.Value{}, false
			}
		} else if i != len(parts)-1 {
			// not last, but already has no deep
			return reflect.Value{}, false
		}
	}

	return cur, true
}

// GetterValue is a special case of FieldValue, that checks if for a field named Foo,
// there is a method that is named FooBytes(). Useful to avoid allocations when accesing
// string values that can be returned as a memory pointer instead (e.g. in capnproto).
func GetterValue(path string, in interface{}) (v interface{}, found bool) {
	defer func() {
		if x := recover(); x != nil {
			return
		}
	}()

	parts := strings.Split(path, ".")
	if len(parts) == 0 {
		return nil, false
	}

	var (
		parent   reflect.Value
		lastPart string
	)

	cur := reflect.ValueOf(in)

	for i, part := range parts {
		m := cur.MethodByName(part)
		typ := m.Type()

		if typ.NumIn() != 0 || typ.NumOut() != 1 {
			continue
		}
		// consider it a getter
		out := m.Call(nil)
		parent = cur
		cur = out[0]

		if cur.NumMethod() == 0 && i != len(parts)-1 {
			// not last, but already has no deep
			return nil, false
		}

		lastPart = part
	}

	if cur.Kind() == reflect.String {
		m := parent.MethodByName(lastPart + "Bytes")
		if m.IsValid() {
			out := m.Call(nil)
			if vv := out[0]; vv.CanInterface() {
				v = vv.Interface()
				found = true

				return
			}
		}
	}
	v = cur.Interface()
	found = true

	return
}

// FieldListNoSort return the list of fields of a struct, recursively without sorting.
func FieldListNoSort(in interface{}) []string {
	defer func() {
		if x := recover(); x != nil {
			return
		}
	}()

	t := reflect.TypeOf(in)
	v := reflect.ValueOf(in)

	for {
		if t.Kind() == reflect.Ptr ||
			t.Kind() == reflect.Interface {
			t = t.Elem()
			v = v.Elem()

			continue
		}

		break
	}

	var flatList []string

	if t.Kind() == reflect.Struct {
		flatList = make([]string, 0, t.NumField())
		flatList = traverseFields("", flatList, t, v)
	} else if t.Kind() == reflect.Map {
		flatList = make([]string, 0, len(v.MapKeys()))
		flatList = traverseMap("", flatList, t, v)
	}

	return flatList
}

// FieldList returns the list of fields of a struct, sorted.
func FieldList(in interface{}) []string {
	flatList := FieldListNoSort(in)

	sort.Strings(flatList)

	return flatList
}

func traverseFields(prefix string, flatList []string, t reflect.Type, v reflect.Value) []string {
	n := t.NumField()
	for i := 0; i < n; i++ {
		var field reflect.Value
		if v.IsValid() {
			field = v.Field(i)
		}

		fieldType := t.Field(i).Type

		for {
			if fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Interface {
				fieldType = fieldType.Elem()

				if field.IsValid() {
					field = field.Elem()
				}
				continue
			}
			break
		}

		fieldPrefix := t.Field(i).Name

		if len(prefix) > 0 {
			fieldPrefix = fmt.Sprintf("%s.%s", prefix, fieldPrefix)
		}

		if fieldType.Kind() == reflect.Struct {
			flatList = traverseFields(fieldPrefix, flatList, fieldType, field)
			continue
		} else if fieldType.Kind() == reflect.Map {
			flatList = traverseMap(fieldPrefix, flatList, fieldType, field)
			continue
		}

		flatList = append(flatList, fieldPrefix)
	}
	return flatList
}

func traverseMap(prefix string, flatList []string, t reflect.Type, v reflect.Value) []string {
	for _, key := range v.MapKeys() {
		var field reflect.Value
		if v.IsValid() {
			field = v.MapIndex(key)
		}
		fieldType := field.Type()

		if (fieldType.Kind() == reflect.Ptr ||
			fieldType.Kind() == reflect.Interface) && !field.IsNil() {
			for {
				if fieldType.Kind() == reflect.Ptr || fieldType.Kind() == reflect.Interface {
					field = field.Elem()
					fieldType = field.Type()
					continue
				}
				break
			}
		}
		fieldPrefix := key.String()
		if len(prefix) > 0 {
			fieldPrefix = fmt.Sprintf("%s.%s", prefix, key.String())
		}

		if fieldType.Kind() == reflect.Struct {
			flatList = traverseFields(fieldPrefix, flatList, fieldType, field)
			continue
		} else if fieldType.Kind() == reflect.Map {
			flatList = traverseMap(fieldPrefix, flatList, fieldType, field)
			continue
		}

		flatList = append(flatList, fieldPrefix)
	}
	return flatList
}

// GetterList returns list of getter methods that accept
// no arguments (except the implicit pointer to the struct) and return one value.
//
// Example: (f *foo) Bar() string
func GetterList(in interface{}) []string {
	defer func() {
		if x := recover(); x != nil {
			return
		}
	}()

	t := reflect.TypeOf(in)
	flatList := make([]string, 0, t.NumMethod())
	flatList = traverseGetters("", flatList, t, reflect.ValueOf(in))
	sort.Strings(flatList)
	return flatList
}

func traverseGetters(prefix string, flatList []string,
	t reflect.Type, v reflect.Value) []string {
	n := t.NumMethod()
	for i := 0; i < n; i++ {
		m := t.Method(i).Type
		if m.NumIn() != 1 || m.NumOut() != 1 {
			continue
		}
		mPrefix := t.Method(i).Name
		if len(prefix) > 0 {
			mPrefix = fmt.Sprintf("%s.%s", prefix, mPrefix)
		}

		out := v.Method(i).Call(nil)
		if out[0].Kind() == reflect.Struct {
			flatList = traverseGetters(mPrefix, flatList, out[0].Type(), out[0])
			continue
		}

		flatList = append(flatList, mPrefix)
	}
	return flatList
}
