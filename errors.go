package config

import "fmt"

// Errors is a list of errors that occurred during processing.
// Combined with IndexErrors and FieldErrors it is possible to narrow down
// faulty input and help a user figure out what they need to fix.
type Errors struct {
	message string
	errs    []error
}

func (e Errors) Error() string {
	return fmt.Sprintf("%s due to %d errors", e.message, len(e.errs))
}

// Errors returns the list of errors that caused this error to be returned
func (e Errors) Errors() []error {
	return e.errs
}

func (e *Errors) appendError(err error) {
	e.errs = append(e.errs, err)
}

func (e *Errors) returnErrors(message string) error {
	if len(e.errs) > 0 {
		e.message = message
		return e
	}
	return nil
}

// IndexError is used when the error occurs while processing a slice/array of input
// values and will tell you which index in the input data (not the output data) to look.
type IndexError struct {
	error
	index int
}

func (e IndexError) Error() string {
	return fmt.Sprintf("%v for index %d", e.error, e.index)
}

// Err returns the original error that this object is wrapping
func (e IndexError) Err() error {
	return e.error
}

// Index of the error in the input slice
func (e IndexError) Index() int {
	return e.index
}

// FieldError is used when the error occurs while processing structural input
// values and will tell you which key in the input it had issues with.
// This could be a recursive structure containing errors all the way down.
type FieldError struct {
	error
	field string
}

func (e FieldError) Error() string {
	return fmt.Sprintf("%v for field %v", e.error, e.field)
}

// Err returns the original error that this object is wrapping
func (e FieldError) Err() error {
	return e.error
}

// Field that has issues.
func (e FieldError) Field() string {
	return e.field
}
