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
// Unset values remain intact or as their native zero value: https://tour.golang.org/basics/12.
//
// Nested structs/subconfigs are delimited with double underscore.
//   PARENT__CHILD
//
// Env vars map to struct fields case insensitively.
// NOTE: Also true when using struct tags.
package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
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
	failedFields            []string
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
// It returns an error if:
//     * struct contains unsupported fields (pointers, maps, slice of structs, channels, arrays, funcs, interfaces, complex)
//     * there were errors doing file i/o
// It panics if:
//     * target is not a struct pointer
func (c *Builder) To(target interface{}) error {
	if reflect.ValueOf(target).Kind() != reflect.Ptr || reflect.ValueOf(target).Elem().Kind() != reflect.Struct {
		panic("config: To(target) must be a *struct")
	}
	c.populateStructRecursively(target, "")
	if c.failedFields != nil {
		return fmt.Errorf("config: the following fields had errors: %v", c.failedFields)
	}
	return nil
}

// From returns a new Builder, populated with the values from file.
func From(file string) *Builder {
	return newBuilder().From(file)
}

// From merges new values from file into the current config state, returning the Builder.
func (c *Builder) From(file string) *Builder {
	content, err := ioutil.ReadFile(file)
	if err != nil {
		c.failedFields = append(c.failedFields, fmt.Sprintf("file[%v]", file))
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var ss []string
	for scanner.Scan() {
		ss = append(ss, scanner.Text())
	}
	if scanner.Err() != nil {
		c.failedFields = append(c.failedFields, fmt.Sprintf("file[%v]", file))
	}
	c.mergeConfig(stringsToMap(ss))
	return c
}

// FromEnv returns a new Builder, populated with environment variables
func FromEnv() *Builder {
	return newBuilder().FromEnv()
}

// FromEnv merges new values from the environment into the current config state, returning the Builder.
func (c *Builder) FromEnv() *Builder {
	c.mergeConfig(stringsToMap(os.Environ()))
	return c
}

func (c *Builder) mergeConfig(in map[string]string) {
	for k, v := range in {
		c.configMap[k] = v
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
//
// failed fields are added to the builder for error reporting
func (c *Builder) populateStructRecursively(structPtr interface{}, prefix string) {
	structValue := reflect.ValueOf(structPtr).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structValue.Type().Field(i)
		fieldPtr := structValue.Field(i).Addr().Interface()

		key := getKey(fieldType, prefix)
		value := c.configMap[key]

		switch fieldType.Type.Kind() {
		case reflect.Struct:
			c.populateStructRecursively(fieldPtr, key+c.structDelim)
		case reflect.Slice:
			for _, index := range convertAndSetSlice(fieldPtr, stringToSlice(value, c.sliceDelim)) {
				c.failedFields = append(c.failedFields, fmt.Sprintf("%v[%v]", key, index))
			}
		default:
			if !convertAndSetValue(fieldPtr, value) {
				c.failedFields = append(c.failedFields, key)
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
// All values are attempted.
// Returns the indices of failed values
func convertAndSetSlice(slicePtr interface{}, values []string) []int {
	sliceVal := reflect.ValueOf(slicePtr).Elem()
	elemType := sliceVal.Type().Elem()

	var failedIndices []int
	for i, s := range values {
		valuePtr := reflect.New(elemType)
		if !convertAndSetValue(valuePtr.Interface(), s) {
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
func convertAndSetValue(settable interface{}, s string) bool {
	if s == "" {
		return true
	}
	settableValue := reflect.ValueOf(settable).Elem()
	var (
		err error
		i   int64
		u   uint64
		b   bool
		f   float64
	)
	switch settableValue.Kind() {
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
		settableValue.SetUint(u)
	case reflect.Bool:
		b, err = strconv.ParseBool(s)
		settableValue.SetBool(b)
	case reflect.Float32, reflect.Float64:
		f, err = strconv.ParseFloat(s, settableValue.Type().Bits())
		settableValue.SetFloat(f)
	default:
		err = fmt.Errorf("config: cannot handle kind %v", settableValue.Type().Kind())
	}
	return err == nil
}
