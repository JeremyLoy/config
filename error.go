package config

import (
	"fmt"
	"os"
)

type Error struct {
	fileErrors  []error
	fieldErrors []error
}

func (e *Error) Error() string {
	var all []error
	for _, err := range e.fileErrors {
		all = append(all, err)
	}
	for _, err := range e.fieldErrors {
		all = append(all, err)
	}
	return fmt.Sprintf("%v", all)
}

func (e *Error) FileNotExistErrors() bool {
	if e == nil {
		return false
	}
	for _, err := range e.fileErrors {
		if os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func (e *Error) FileParseErrors() bool {
	if e == nil {
		return false
	}
	for _, err := range e.fileErrors {
		if !os.IsNotExist(err) {
			return true
		}
	}
	return false
}

func (e *Error) FieldParseErrors() bool {
	if e == nil {
		return false
	}
	return len(e.fieldErrors) > 0
}

func (b *Builder) appendFileError(err error) {
	if b.err == nil {
		b.err = &Error{}
	}
	b.err.fileErrors = append(b.err.fileErrors, err)
}

func (b *Builder) appendFieldError(err error, fieldName string, fieldType string) {
	if b.err == nil {
		b.err = &Error{}
	}
	b.err.fieldErrors = append(b.err.fieldErrors, err)
}

func (b *Builder) appendSliceError(err error, fieldName, fieldType string, index int) {
	if b.err == nil {
		b.err = &Error{}
	}
	b.err.fieldErrors = append(b.err.fieldErrors, &sliceError{fieldName, fieldType, index, err})
}

type fieldError struct {
	name string
	t    string
	err  error
}

func (e *fieldError) Error() string {
	return fmt.Sprintf("failed to parse %v value for field %v: %v", e.t, e.name, e.err)
}

type sliceError struct {
	name  string
	t     string
	index int
	err   error
}

func (e *sliceError) Error() string {
	return fmt.Sprintf("failed to parse %v value for slice %v at index %v: %v", e.t, e.name, e.index, e.err)
}
