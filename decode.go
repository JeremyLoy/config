package config

import (
	"fmt"
	"reflect"
	"sort"
)

func (c *Builder) decode(target interface{}, prefix string) error {
	structPtr := reflect.ValueOf(target)
	if structPtr.Kind() != reflect.Ptr || structPtr.Elem().Kind() != reflect.Struct {
		panic("config: To(target) must be a *struct")
	}
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
				c.failedFields = append(c.failedFields, fmt.Sprintf("%v[%v]", key, index))
			}
		default:
			if !convertAndSetValue(fieldPtr, stringValue) {
				c.failedFields = append(c.failedFields, key)
			}
		}
	}
	sort.Strings(c.failedFields) // make output deterministic to aid in debugging
	if c.failedFields != nil {
		return fmt.Errorf("config: the following fields had errors: %v", c.failedFields)
	}
	return nil
}
