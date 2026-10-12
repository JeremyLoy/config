package config

import (
	"fmt"
	"reflect"
	"slices"
)

func (c *Builder) decode(target any, prefix string) error {
	structPtr := reflect.ValueOf(target)
	if structPtr.Kind() != reflect.Pointer || structPtr.Elem().Kind() != reflect.Struct {
		panic("config: To(target) must be a *struct")
	}
	// Start from the Builder's source errors, but keep field errors local to
	// this call so that repeated calls to To/Sub do not accumulate duplicates.
	failedFields := append([]string(nil), c.failedFields...)
	m := make(map[string]reflect.Value)
	mapKeysToFields(structPtr, m, prefix)

	for key, fieldPtr := range m {
		stringValue, ok := c.configMap[key]
		if !ok {
			continue
		}
		switch fieldPtr.Elem().Type().Kind() {
		case reflect.Slice:
			for _, index := range convertAndSetSlice(fieldPtr, stringToSlice(stringValue, c.sliceDelim)) {
				failedFields = append(failedFields, fmt.Sprintf("%v[%v]", key, index))
			}
		default:
			if !convertAndSetValue(fieldPtr, stringValue) {
				failedFields = append(failedFields, key)
			}
		}
	}
	slices.Sort(failedFields) // make output deterministic to aid in debugging
	if failedFields != nil {
		return fmt.Errorf("config: the following fields had errors: %v", failedFields)
	}
	return nil
}
