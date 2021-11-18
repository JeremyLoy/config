package config

import (
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"
)

func mapKeysToFields(ptr reflect.Value, m map[string]reflect.Value, prefix string) {
	structValue := ptr.Elem()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structValue.Type().Field(i)
		fieldPtr := structValue.Field(i).Addr()

		key := getKey(fieldType, prefix)

		switch fieldType.Type.Kind() {
		case reflect.Struct:
			mapKeysToFields(fieldPtr, m, key+structDelim)
		default:
			m[key] = fieldPtr
		}
	}
}

func mergeMaps(first, second map[string]string) {
	for k, v := range second {
		first[k] = v
	}
}

// stringsToMap builds a map from a string slice.
// The input strings are assumed to be environment variable in style e.g. KEY=VALUE
// Keys with no value are not added to the map.
func stringsToMap(ss []string) map[string]string {
	m := make(map[string]string)
	for _, s := range ss {
		if !strings.Contains(s, "=") {
			continue // ensures return is always of length 2
		}
		// TODO replace with strings.Cut in go 1.18
		split := strings.SplitN(s, "=", 2)
		key, value := strings.ToLower(split[0]), split[1]
		if key != "" && value != "" {
			m[key] = value
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
// It converts each entry in "values" to the elemType of the passed in slice.
// The slice remains nil if "values" is empty.
// All values are attempted.
// Returns the indices of failed values
func convertAndSetSlice(slicePtr reflect.Value, values []string) []int {
	sliceVal := slicePtr.Elem()
	elemType := sliceVal.Type().Elem()

	var failedIndices []int
	for i, s := range values {
		valuePtr := reflect.New(elemType)
		if !convertAndSetValue(valuePtr, s) {
			failedIndices = append(failedIndices, i)
		} else {
			sliceVal.Set(reflect.Append(sliceVal, valuePtr.Elem()))
		}
	}
	return failedIndices
}

// convertAndSetValue receives a settable of an arbitrary kind, and sets its value to s, returning true.
// It calls the matching strconv function on s, based on the settable's kind.
// All basic types (bool, int, float, string) are handled by this function.
// Slice and struct are handled elsewhere.
//
// An unhandled kind or a failed parse returns false.
// False is used to prevent accidental logging of secrets as
// as the strconv include s in their error message.
func convertAndSetValue(settable reflect.Value, s string) bool {
	settableValue := settable.Elem()
	var (
		err error
		i   int64
		u   uint64
		b   bool
		f   float64
		url *url.URL
	)
	switch settableValue.Kind() {
	case reflect.Ptr: // only one pointer type is handled at the moment, *url.URL
		url, err = url.Parse(s)
		if err == nil {
			settableValue.Set(reflect.ValueOf(url))
		}
	case reflect.String:
		settableValue.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if settableValue.Type().PkgPath() == "time" && settableValue.Type().Name() == "Duration" {
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
		err = fmt.Errorf("config: cannot handle kind %v", settableValue.Type().Kind())
	}
	return err == nil
}
