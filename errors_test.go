package config_test

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/JeremyLoy/config"
)

func Test_typedErrors(t *testing.T) {
	t.Parallel()
	var c struct {
		Port    int8
		Secret  int
		Ptr     *int
		Timeout bool
	}
	missing := filepath.Join(t.TempDir(), "missing")
	file := writeConfigFile(t, "PORT=128\nSECRET=hunter2\nPTR=1\nTIMEOUT=maybe")
	err := config.From(file).From(missing).To(&c)

	cfgErr, ok := errors.AsType[*config.Error](err)
	if !ok {
		t.Fatalf("got %T, want *config.Error", err)
	}
	var keys []string
	for _, fe := range cfgErr.Errors {
		keys = append(keys, fe.Key)
	}
	wantKeys := "file[" + missing + "] port ptr secret timeout"
	if got := strings.Join(keys, " "); got != wantKeys {
		t.Errorf("got keys %q, want %q", got, wantKeys)
	}

	for _, target := range []error{fs.ErrNotExist, strconv.ErrRange, strconv.ErrSyntax, config.ErrUnsupportedType} {
		if !errors.Is(err, target) {
			t.Errorf("errors.Is(err, %v) = false, want true", target)
		}
	}
	for _, fe := range cfgErr.Errors {
		if strings.Contains(fe.Error(), "hunter2") {
			t.Errorf("FieldError %q leaks the configured value", fe.Error())
		}
	}
}

func Example_errorInspection() {
	os.Clearenv()
	os.Setenv("PORT", "256")

	var c struct{ Port uint8 }
	err := config.FromEnv().To(&c)

	if cfgErr, ok := errors.AsType[*config.Error](err); ok {
		for _, fe := range cfgErr.Errors {
			fmt.Println(fe)
		}
	}
	fmt.Println(errors.Is(err, strconv.ErrRange))

	// Output:
	// port: value out of range
	// true
}
