// Package config provides typesafe, cloud native configuration binding from environment variables or files to structs.
//
// Configuration can be done in as little as two lines:
//     var c MyConfig
//     config.FromEnv().To(&c)
//
// A field's type determines what https://golang.org/pkg/strconv/ function is called.
//
// All string conversion rules are as defined in the https://golang.org/pkg/strconv/ package.
//
// If chaining multiple data sources, data sets are merged.
//
// Later values override previous values.
//   config.From("dev.config").FromEnv().To(&c)
//
// Unset values remain as their native zero value: https://tour.golang.org/basics/12.
//
// Nested structs/subconfigs are delimited with double underscore.
//   PARENT__CHILD
//
// Env vars map to struct fields case insensitively.
// NOTE: Also true when using struct tags.
package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const (
	structTagKey = "config"
	structDelim  = "__"
	sliceDelim   = " "
)

// Builder contains the current configuration state.
type Builder struct {
	structDelim, sliceDelim string
	configMap               map[string]string
	err                     *Error
}

func newBuilder() *Builder {
	return &Builder{
		configMap:   make(map[string]string),
		structDelim: structDelim,
		sliceDelim:  sliceDelim,
	}
}

// To accepts a struct pointer, and populates it with the current config state.
// Supported fields:
//     * all int, uint, float variants
//     * bool, struct, string
//     * slice of any of the above, except for []struct{}
// It panics under the following circumstances:
//     * target is not a struct pointer
//     * struct contains unsupported fields (pointers, maps, slice of structs, channels, arrays, funcs, interfaces, complex)
// It either returns nil or a *Error, aggregating all errors encountered by Builder
func (b *Builder) To(target interface{}) error {
	b.populateStructRecursively(target, "")
	return b.err
}

// From returns a new Builder, populated with the values from file.
// It panics if unable to open the file.
func From(file string) *Builder {
	return newBuilder().From(file)
}

// From merges new values from file into the current config state, returning the Builder.
// It panics if unable to open the file.
func (b *Builder) From(file string) *Builder {
	f, err := os.Open(file)
	if err != nil {
		b.appendFileError(err)
		return b
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	var ss []string
	for scanner.Scan() {
		ss = append(ss, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		b.appendFileError(err)
	}
	b.mergeConfig(stringsToMap(ss))
	return b
}

// FromEnv returns a new Builder, populated with environment variables
func FromEnv() *Builder {
	return newBuilder().FromEnv()
}

// FromEnv merges new values from the environment into the current config state, returning the Builder.
func (b *Builder) FromEnv() *Builder {
	b.mergeConfig(stringsToMap(os.Environ()))
	return b
}

func (b *Builder) mergeConfig(in map[string]string) {
	for k, v := range in {
		b.configMap[k] = v
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
		split := strings.SplitN(s, "=", 2)
		key, value := strings.ToLower(split[0]), split[1]
		if key != "" && value != "" {
			m[key] = value
		}
	}
	return m
}

// populateStructRecursively populates each field of the passed in struct.
// slices and values are set directly.
// nested structs recurse through this function.
// values are derived from the field name, prefixed with the field names of any parents.
func (b *Builder) populateStructRecursively(structPtr interface{}, prefix string) {
	structValue := reflect.ValueOf(structPtr).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structValue.Type().Field(i)
		fieldPtr := structValue.Field(i).Addr().Interface()

		key := getKey(fieldType, prefix)
		value := b.configMap[key]

		switch fieldType.Type.Kind() {
		case reflect.Struct:
			b.populateStructRecursively(fieldPtr, key+b.structDelim)
		case reflect.Slice:
			b.convertAndSetSlice(fieldPtr, stringToSlice(value, b.sliceDelim), key)
		default:
			if err := convertAndSetValue(fieldPtr, value); err != nil {
				b.appendFieldError(err, key, fieldType.Type.String())
			}
		}
	}
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
func (b *Builder) convertAndSetSlice(slicePtr interface{}, values []string, fieldName string) {
	sliceVal := reflect.ValueOf(slicePtr).Elem()
	elemType := sliceVal.Type().Elem()

	for i, s := range values {
		valuePtr := reflect.New(elemType)
		if err := convertAndSetValue(valuePtr.Interface(), s); err != nil {
			b.appendSliceError(err, fieldName, elemType.String(), i)
		} else {
			sliceVal.Set(reflect.Append(sliceVal, valuePtr.Elem()))
		}
	}
}

// convertAndSetValue receives a settable of an arbitrary kind, and sets its value to s.
// It calls the matching strconv function on s, based on the settable's kind.
// All basic types (bool, int, float, string) are handled by this function.
// Slice and struct are handled elsewhere.
// Unhandled kinds panic.
// Errors in string conversion are ignored, and the settable remains a zero value.
func convertAndSetValue(settable interface{}, s string) (err error) {
	settableValue := reflect.ValueOf(settable).Elem()
	switch settableValue.Kind() {
	case reflect.String:
		settableValue.SetString(s)
	case reflect.Int:
		var val int64
		if val, err = strconv.ParseInt(s, 10, 0); err == nil {
			settableValue.SetInt(val)
		}
	case reflect.Int8:
		var val int64
		if val, err = strconv.ParseInt(s, 10, 8); err == nil {
			settableValue.SetInt(val)
		}
	case reflect.Int16:
		var val int64
		if val, err = strconv.ParseInt(s, 10, 16); err == nil {
			settableValue.SetInt(val)
		}
	case reflect.Int32:
		var val int64
		if val, err = strconv.ParseInt(s, 10, 32); err == nil {
			settableValue.SetInt(val)
		}
	case reflect.Int64:
		var val int64
		if val, err = strconv.ParseInt(s, 10, 64); err == nil {
			settableValue.SetInt(val)
		}
	case reflect.Uint:
		var val uint64
		if val, err = strconv.ParseUint(s, 10, 0); err == nil {
			settableValue.SetUint(val)
		}
	case reflect.Uint8:
		var val uint64
		if val, err = strconv.ParseUint(s, 10, 8); err == nil {
			settableValue.SetUint(val)
		}
	case reflect.Uint16:
		var val uint64
		if val, err = strconv.ParseUint(s, 10, 16); err == nil {
			settableValue.SetUint(val)
		}
	case reflect.Uint32:
		var val uint64
		if val, err = strconv.ParseUint(s, 10, 32); err == nil {
			settableValue.SetUint(val)
		}
	case reflect.Uint64:
		var val uint64
		if val, err = strconv.ParseUint(s, 10, 64); err == nil {
			settableValue.SetUint(val)
		}
	case reflect.Bool:
		var val bool
		if val, err = strconv.ParseBool(s); err == nil {
			settableValue.SetBool(val)
		}
	case reflect.Float32:
		var val float64
		if val, err = strconv.ParseFloat(s, 32); err == nil {
			settableValue.SetFloat(val)
		}
	case reflect.Float64:
		var val float64
		if val, err = strconv.ParseFloat(s, 64); err == nil {
			settableValue.SetFloat(val)
		}
	default:
		panic(fmt.Sprintf("cannot handle kind %v\n", settableValue.Type().Kind()))
	}
	return
}
