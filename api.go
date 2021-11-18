package config

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
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

// To accepts a struct pointer, and populates it with the current config state.
// Supported fields:
//     * all int, uint, float variants
//     * bool, struct, string
//     * time.Duration
//     * \*url.URL
//     * slice of any of the above, except for []struct{}
// It returns an error if:
//     * struct contains unsupported fields (pointers, maps, slice of structs, channels, arrays, funcs, interfaces, complex)
//     * there were errors doing file i/o
// It panics if:
//     * target is not a struct pointer
func (c *Builder) To(target interface{}) error {
	return c.decode(target, "")
}

// Sub behaves the same as To, however it maps configuration to the struct starting at the given prefix.
func (c *Builder) Sub(target interface{}, prefix string) error {
	return c.decode(target, prefix+c.structDelim)
}

// From returns a new Builder, populated with the values from file.
func From(file string) *Builder {
	return newBuilder().From(file)
}

// FromOptional returns a new Builder, populated with the values from file if it exists.
func FromOptional(file string) *Builder {
	return newBuilder().FromOptional(file)
}

// From merges new values from file into the current config state, returning the Builder.
func (c *Builder) From(file string) *Builder {
	return c.appendFile(file, true)
}

// FromOptional merges new values from file (if it exists) into the current config state, returning the Builder.
func (c *Builder) FromOptional(file string) *Builder {
	return c.appendFile(file, false)
}

func (c *Builder) appendFile(file string, includeErr bool) *Builder {
	content, err := os.ReadFile(file)
	if includeErr && err != nil {
		c.failedFields = append(c.failedFields, fmt.Sprintf("file[%v]", file))
	}
	scanner := bufio.NewScanner(bytes.NewReader(content))
	var ss []string
	for scanner.Scan() {
		ss = append(ss, scanner.Text())
	}
	if includeErr && scanner.Err() != nil {
		c.failedFields = append(c.failedFields, fmt.Sprintf("file[%v]", file))
	}
	mergeMaps(c.configMap, stringsToMap(ss))
	return c
}

// FromEnv returns a new Builder, populated with environment variables
func FromEnv() *Builder {
	return newBuilder().FromEnv()
}

// FromEnv merges new values from the environment into the current config state, returning the Builder.
func (c *Builder) FromEnv() *Builder {
	mergeMaps(c.configMap, stringsToMap(os.Environ()))
	return c
}

func newBuilder() *Builder {
	return &Builder{
		configMap:   make(map[string]string),
		structDelim: structDelim,
		sliceDelim:  sliceDelim,
	}
}
