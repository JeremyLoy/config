package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"
)

func (c *Builder) decode(target any, prefix string) error {
	structPtr := reflect.ValueOf(target)
	if structPtr.Kind() != reflect.Pointer || structPtr.Elem().Kind() != reflect.Struct {
		panic("config: To(target) must be a *struct")
	}
	// Start from the Builder's source errors, but keep field errors local to
	// this call so that repeated calls to To/Sub do not accumulate duplicates.
	errs := slices.Clone(c.sourceErrs)
	m := make(map[string]reflect.Value)
	mapKeysToFields(structPtr, m, prefix)

	for key, fieldPtr := range m {
		stringValue, ok := c.configMap[key]
		if !ok {
			continue
		}
		switch fieldType := fieldPtr.Elem().Type(); {
		case fieldType.Kind() == reflect.Slice && !isTextUnmarshaler(fieldType):
			for index, err := range convertAndSetSlice(fieldPtr, stringToSlice(stringValue, sliceDelim)) {
				errs = append(errs, &FieldError{Key: fmt.Sprintf("%v[%v]", key, index), Err: err})
			}
		default:
			if err := convertAndSetValue(fieldPtr, stringValue); err != nil {
				errs = append(errs, &FieldError{Key: key, Err: err})
			}
		}
	}
	if errs != nil {
		// make output deterministic to aid in debugging
		slices.SortFunc(errs, func(a, b *FieldError) int { return strings.Compare(a.Key, b.Key) })
		return &Error{Errors: errs}
	}
	return nil
}
