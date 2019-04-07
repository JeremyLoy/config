package config

import (
	"errors"
	"os"
	"strconv"
	"testing"
)

func Test_fieldError_Error(t *testing.T) {
	type fields struct {
		name string
		t    string
		err  error
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "expected use",
			fields: fields{
				name: "port",
				t:    "int",
				err:  func() error { _, err := strconv.Atoi("xyz"); return err }(),
			},
			want: "failed to parse int value for field port: strconv.Atoi: parsing \"xyz\": invalid syntax",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &fieldError{
				name: tt.fields.name,
				t:    tt.fields.t,
				err:  tt.fields.err,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("fieldError.Error() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_FileNotExistErrors(t *testing.T) {
	t.Parallel()

	err := &Error{
		fileErrors: []error{os.ErrNotExist},
	}

	if !err.FileNotExistErrors() {
		t.Errorf("expected file not found error")
	}
	if err.FileParseErrors() {
		t.Errorf("expected to not have general file erros")
	}
}

func TestError_FileErrors(t *testing.T) {
	t.Parallel()

	err := &Error{
		fileErrors: []error{errors.New("oops")},
	}

	if err.FileNotExistErrors() {
		t.Errorf("expected not to have file not found error")
	}
	if !err.FileParseErrors() {
		t.Errorf("expected to have general file erros")
	}
}
