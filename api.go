package config

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"reflect"
	"strings"
)

const (
	structTagKey = "config"
	structDelim  = "__"
	sliceDelim   = " "
)

// Builder contains the current configuration state.
type Builder struct {
	configMap  map[string]string
	sourceErrs []*FieldError
}

// To accepts a struct pointer, and populates it with the current config state.
// Supported fields:
//   - all int, uint, float variants
//   - bool, struct, string
//   - time.Duration
//   - \*url.URL
//   - any type whose pointer implements [encoding.TextUnmarshaler],
//     such as [time.Time], [net/netip.Addr], [net.IP] and [log/slog.Level]
//   - slice of any of the above, except for []struct{}
//
// Unexported fields are ignored. Slices are replaced, not appended to.
//
// It returns an *Error if:
//   - struct contains unsupported fields (pointers other than *url.URL, maps, slice of structs, channels, arrays, funcs, interfaces, complex)
//   - there were errors doing file i/o
//
// It panics if:
//   - target is not a struct pointer
func (c *Builder) To(target any) error {
	return c.decode(mustStructPtr(target), "")
}

// Sub behaves the same as To, however it maps configuration to the struct starting at the given prefix.
func (c *Builder) Sub(target any, prefix string) error {
	return c.decode(mustStructPtr(target), prefix+structDelim)
}

// Load returns a new T populated with the current config state.
// It behaves the same as To, except that it returns ErrInvalidTarget
// instead of panicking if T is not a struct type.
//
//	cfg, err := config.From("dev.config").FromEnv().Load[MyConfig]()
func (c *Builder) Load[T any]() (T, error) {
	var target T
	if reflect.TypeFor[T]().Kind() != reflect.Struct {
		return target, ErrInvalidTarget
	}
	err := c.decode(reflect.ValueOf(&target), "")
	return target, err
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
	return c.appendFile(os.ReadFile, file, false)
}

// FromOptional merges new values from file (if it exists) into the current config state, returning the Builder.
func (c *Builder) FromOptional(file string) *Builder {
	return c.appendFile(os.ReadFile, file, true)
}

// FromFS returns a new Builder, populated with the values from the named file in fsys.
func FromFS(fsys fs.FS, name string) *Builder {
	return newBuilder().FromFS(fsys, name)
}

// FromFS merges new values from the named file in fsys into the current config state, returning the Builder.
func (c *Builder) FromFS(fsys fs.FS, name string) *Builder {
	return c.appendFile(func(name string) ([]byte, error) { return fs.ReadFile(fsys, name) }, name, false)
}

// FromReader returns a new Builder, populated with the values read from r.
func FromReader(r io.Reader) *Builder {
	return newBuilder().FromReader(r)
}

// FromReader merges new values read from r into the current config state, returning the Builder.
// r uses the same KEY=VALUE format as a file.
func (c *Builder) FromReader(r io.Reader) *Builder {
	return c.appendReader(r, "reader")
}

// FromMap returns a new Builder, populated with the values in m.
func FromMap(m map[string]string) *Builder {
	return newBuilder().FromMap(m)
}

// FromMap merges the values in m into the current config state, returning the Builder.
// Keys are matched case insensitively, and empty values are ignored, as with every other source.
func (c *Builder) FromMap(m map[string]string) *Builder {
	for k, v := range m {
		if k != "" && v != "" {
			c.configMap[strings.ToLower(k)] = v
		}
	}
	return c
}

// appendFile merges the contents of file, as returned by readFile, into the config state.
// If optional is true, a file that does not exist is not an error; any other failure still is.
func (c *Builder) appendFile(readFile func(string) ([]byte, error), file string, optional bool) *Builder {
	key := fmt.Sprintf("file[%v]", file)
	content, err := readFile(file)
	if err != nil {
		if !(optional && errors.Is(err, fs.ErrNotExist)) {
			c.sourceErrs = append(c.sourceErrs, &FieldError{Key: key, Err: err})
		}
		return c
	}
	return c.appendReader(bytes.NewReader(content), key)
}

// appendReader merges KEY=VALUE lines read from r into the config state.
// Lines starting with # are comments and are ignored. key identifies r in errors.
func (c *Builder) appendReader(r io.Reader, key string) *Builder {
	scanner := bufio.NewScanner(r)
	var ss []string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(strings.TrimSpace(line), "#") {
			continue
		}
		ss = append(ss, line)
	}
	if err := scanner.Err(); err != nil {
		c.sourceErrs = append(c.sourceErrs, &FieldError{Key: key, Err: err})
	}
	maps.Copy(c.configMap, stringsToMap(ss))
	return c
}

// FromEnv returns a new Builder, populated with environment variables
func FromEnv() *Builder {
	return newBuilder().FromEnv()
}

// FromEnv merges new values from the environment into the current config state, returning the Builder.
func (c *Builder) FromEnv() *Builder {
	maps.Copy(c.configMap, stringsToMap(os.Environ()))
	return c
}

func newBuilder() *Builder {
	return &Builder{configMap: make(map[string]string)}
}
