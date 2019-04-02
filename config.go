package config

import (
	"bufio"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

const delim = "__"

type ConfigBuilder interface {
	FromEnv() ConfigBuilder
	From(file string) ConfigBuilder
	To(target interface{})
}

type configBuilder struct {
	delim     string
	configMap map[string]string
}

func newConfigBuilder() ConfigBuilder {
	return &configBuilder{
		configMap: make(map[string]string),
		delim:     delim,
	}
}

func (c *configBuilder) mergeConfig(in map[string]string) {
	for k, v := range in {
		c.configMap[k] = v
	}
}

func From(file string) ConfigBuilder {
	return newConfigBuilder().From(file)
}

func (c *configBuilder) From(f string) ConfigBuilder {
	file, err := os.Open(f)
	if err != nil {
		panic("oops!")
	}
	scanner := bufio.NewScanner(file)
	var ss []string
	for scanner.Scan() {
		ss = append(ss, scanner.Text())
	}
	c.mergeConfig(stringsToMap(ss))
	return c
}

func FromEnv() ConfigBuilder {
	return newConfigBuilder().FromEnv()
}

func stringsToMap(ss []string) map[string]string {
	m := make(map[string]string)
	for _, s := range ss {
		kv := strings.SplitN(s, "=", 2)
		m[kv[0]] = kv[1]
	}
	return m
}

func (c *configBuilder) FromEnv() ConfigBuilder {
	c.mergeConfig(stringsToMap(os.Environ()))
	return c
}

func (c *configBuilder) To(target interface{}) {
	c.populateStructRecursively(target, "")
}

func (c *configBuilder) populateStructRecursively(s interface{}, prefix string) {
	structValue := reflect.ValueOf(s).Elem()
	for i := 0; i < structValue.NumField(); i++ {
		fieldType := structValue.Type().Field(i)
		fieldValue := structValue.Field(i)

		kind := fieldType.Type.Kind()

		key := prefix + fieldType.Name
		value, _ := c.configMap[key]

		// If a ptr, initialize with a new zero value
		// Set fieldValue and kind to the underlying elem and kind.
		if kind == reflect.Ptr {
			fieldValue.Set(reflect.New(fieldType.Type.Elem()))
			fieldValue = fieldValue.Elem()
			kind = fieldType.Type.Elem().Kind()
		}

		switch kind {
		case reflect.Struct:
			// Recurse, passing in the address of this value (struct) and setting the prefix to be the current key + delim.
			c.populateStructRecursively(fieldValue.Addr().Interface(), key+c.delim)
		case reflect.String:
			fieldValue.SetString(value)
		case reflect.Int:
			val, _ := strconv.ParseInt(value, 10, 0)
			fieldValue.SetInt(val)
		case reflect.Int8:
			val, _ := strconv.ParseInt(value, 10, 8)
			fieldValue.SetInt(val)
		case reflect.Int16:
			val, _ := strconv.ParseInt(value, 10, 26)
			fieldValue.SetInt(val)
		case reflect.Int32:
			val, _ := strconv.ParseInt(value, 10, 32)
			fieldValue.SetInt(val)
		case reflect.Int64:
			val, _ := strconv.ParseInt(value, 10, 64)
			fieldValue.SetInt(val)
		case reflect.Uint:
			val, _ := strconv.ParseUint(value, 10, 0)
			fieldValue.SetUint(val)
		case reflect.Uint8:
			val, _ := strconv.ParseUint(value, 10, 8)
			fieldValue.SetUint(val)
		case reflect.Uint16:
			val, _ := strconv.ParseUint(value, 10, 16)
			fieldValue.SetUint(val)
		case reflect.Uint32:
			val, _ := strconv.ParseUint(value, 10, 32)
			fieldValue.SetUint(val)
		case reflect.Uint64:
			val, _ := strconv.ParseUint(value, 10, 64)
			fieldValue.SetUint(val)
		case reflect.Bool:
			val, _ := strconv.ParseBool(value)
			fieldValue.SetBool(val)
		case reflect.Float32:
			val, _ := strconv.ParseFloat(value, 32)
			fieldValue.SetFloat(val)
		case reflect.Float64:
			val, _ := strconv.ParseFloat(value, 64)
			fieldValue.SetFloat(val)
		default:
			panic(fmt.Sprintf("cannot handle kind %v\n", fieldType.Type.Kind()))
		}
	}
}
