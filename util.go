package config

import (
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	// urlPtrType is the only pointer type supported as a field.
	urlPtrType   = reflect.TypeFor[*url.URL]()
	durationType = reflect.TypeFor[time.Duration]()
)

func mapKeysToFields(ptr reflect.Value, m map[string]reflect.Value, prefix string) {
	for fieldType, fieldValue := range ptr.Elem().Fields() {
		if !fieldType.IsExported() {
			continue // unexported fields cannot be set
		}
		fieldPtr := fieldValue.Addr()

		key := getKey(fieldType, prefix)

		switch fieldType.Type.Kind() {
		case reflect.Struct:
			mapKeysToFields(fieldPtr, m, key+structDelim)
		default:
			m[key] = fieldPtr
		}
	}
}

// stringsToMap builds a map from a string slice.
// The input strings are assumed to be environment variable in style e.g. KEY=VALUE
// Keys with no value are not added to the map.
func stringsToMap(ss []string) map[string]string {
	m := make(map[string]string)
	for _, s := range ss {
		key, value, ok := strings.Cut(s, "=")
		if ok && key != "" && value != "" {
			m[strings.ToLower(key)] = value
		}
	}
	return m
}

// getKey returns the string that represents this structField in the config map.
// If the structField has the appropriate structTag set, it is used.
// Otherwise, field's name is used.
func getKey(t reflect.StructField, prefix string) string {
	name := t.Name
	if tag, exists := t.Tag.Lookup(structTagKey); exists {
		if tag = strings.TrimSpace(tag); tag != "" {
			name = tag
		}
	}
	return strings.ToLower(prefix + name)
}

// stringToSlice converts a string to a slice of string, using delim.
// It strips surrounding whitespace of all entries.
// If the input string is empty or all whitespace, nil is returned.
func stringToSlice(s, delim string) []string {
	if delim == "" {
		panic("empty delimiter") // impossible or programmer error
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	split := strings.Split(s, delim)
	filtered := split[:0] // https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	for _, v := range split {
		v = strings.TrimSpace(v)
		if v != "" {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// convertAndSetSlice builds a slice of a dynamic type.
// It converts each entry in "values" to the elemType of the passed in slice,
// and replaces the slice's existing contents with the converted values.
// The slice is left untouched if "values" is empty.
// All values are attempted.
// Returns the errors of failed values, keyed by index.
func convertAndSetSlice(slicePtr reflect.Value, values []string) map[int]error {
	if len(values) == 0 {
		return nil
	}
	sliceVal := slicePtr.Elem()
	elemType := sliceVal.Type().Elem()

	newSlice := reflect.MakeSlice(sliceVal.Type(), 0, len(values))
	var failed map[int]error
	for i, s := range values {
		valuePtr := reflect.New(elemType)
		if err := convertAndSetValue(valuePtr, s); err != nil {
			if failed == nil {
				failed = make(map[int]error)
			}
			failed[i] = err
		} else {
			newSlice = reflect.Append(newSlice, valuePtr.Elem())
		}
	}
	sliceVal.Set(newSlice)
	return failed
}

// convertAndSetValue receives a settable of an arbitrary kind, and sets its value to s.
// It calls the matching strconv function on s, based on the settable's kind.
// All basic types (bool, int, float, string) are handled by this function.
// Slice and struct are handled elsewhere.
//
// An unhandled kind returns ErrUnsupportedType.
// A failed parse returns a sanitized error that does not contain s,
// to prevent accidental logging of secrets.
func convertAndSetValue(settable reflect.Value, s string) error {
	settableValue := settable.Elem()
	var (
		err error
		i   int64
		u   uint64
		b   bool
		f   float64
		p   *url.URL
	)
	switch settableValue.Kind() {
	case reflect.Pointer:
		if settableValue.Type() != urlPtrType {
			return ErrUnsupportedType
		}
		p, err = url.Parse(s)
		if err == nil {
			settableValue.Set(reflect.ValueOf(p))
		}
	case reflect.String:
		settableValue.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if settableValue.Type() == durationType {
			var d time.Duration
			d, err = time.ParseDuration(s)
			i = int64(d)
		} else {
			i, err = strconv.ParseInt(s, 10, settableValue.Type().Bits())
		}
		if err == nil {
			settableValue.SetInt(i)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err = strconv.ParseUint(s, 10, settableValue.Type().Bits())
		if err == nil {
			settableValue.SetUint(u)
		}
	case reflect.Bool:
		b, err = strconv.ParseBool(s)
		if err == nil {
			settableValue.SetBool(b)
		}
	case reflect.Float32, reflect.Float64:
		f, err = strconv.ParseFloat(s, settableValue.Type().Bits())
		if err == nil {
			settableValue.SetFloat(f)
		}
	default:
		return ErrUnsupportedType
	}
	if err != nil {
		return sanitize(err)
	}
	return nil
}
