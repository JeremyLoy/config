package config

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

var (
	// ErrInvalidTarget is returned by Load when its type argument is not a struct type.
	ErrInvalidTarget = errors.New("config: Builder.Load[T] requires a struct type")
	// ErrUnsupportedType is the cause of a FieldError for a field whose type cannot be configured.
	ErrUnsupportedType = errors.New("unsupported field type")
	// ErrInvalidValue is the cause of a FieldError for a value that could not be parsed,
	// when no more specific cause (such as strconv.ErrSyntax or strconv.ErrRange) is available.
	ErrInvalidValue = errors.New("invalid value")
)

// FieldError describes a single field or source that could not be loaded.
type FieldError struct {
	// Key identifies what failed: a lower-cased field key such as "port" or "sub__hosts[1]",
	// or "file[name]" for a file that could not be read.
	Key string
	// Err is the cause. It never contains the configured value, so it is safe to log.
	Err error
}

func (e *FieldError) Error() string { return e.Key + ": " + e.Err.Error() }

func (e *FieldError) Unwrap() error { return e.Err }

// Error is returned by To and Sub when one or more fields or sources failed.
// All fields are still attempted; Errors holds every failure, sorted by Key.
//
// Use errors.AsType[*config.Error] to inspect it, or errors.Is to test for a cause
// such as fs.ErrNotExist or strconv.ErrRange.
type Error struct {
	Errors []*FieldError
}

func (e *Error) Error() string {
	keys := make([]string, len(e.Errors))
	for i, fe := range e.Errors {
		keys[i] = fe.Key
	}
	return fmt.Sprintf("config: the following fields had errors: [%s]", strings.Join(keys, " "))
}

func (e *Error) Unwrap() []error {
	errs := make([]error, len(e.Errors))
	for i, fe := range e.Errors {
		errs[i] = fe
	}
	return errs
}

// sanitize reduces a parse error to a cause that does not contain the parsed value,
// so that secrets are never logged.
func sanitize(err error) error {
	if numErr, ok := errors.AsType[*strconv.NumError](err); ok {
		return numErr.Err
	}
	return ErrInvalidValue
}
