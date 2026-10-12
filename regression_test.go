package config_test

import (
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JeremyLoy/config"
)

// writeConfigFile writes content to a new file in a temporary directory and returns its path.
func writeConfigFile(t *testing.T, content string) string {
	t.Helper()
	name := filepath.Join(t.TempDir(), "config")
	if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
	return name
}

func Test_unsupportedPointerIsAnError(t *testing.T) {
	t.Parallel()
	var c struct{ J *int }
	err := config.From(writeConfigFile(t, "J=1")).To(&c)
	if err == nil || err.Error() != "config: the following fields had errors: [j]" {
		t.Errorf("got err %v, want error for field j", err)
	}
	if c.J != nil {
		t.Errorf("got J %v, want nil", c.J)
	}
}

func Test_urlIsParsedVerbatim(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{"/a/../b", "https://example.com/a/../b?q=1", "relative/path"} {
		var c struct{ U *url.URL }
		if err := config.From(writeConfigFile(t, "U="+raw)).To(&c); err != nil {
			t.Fatalf("%q: unexpected error %v", raw, err)
		}
		if c.U == nil || c.U.String() != raw {
			t.Errorf("got %v, want %q", c.U, raw)
		}
	}
}

func Test_unexportedFieldsAreIgnored(t *testing.T) {
	t.Parallel()
	var c struct {
		x int
		Y int
	}
	if err := config.From(writeConfigFile(t, "X=1\nY=2")).To(&c); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if c.x != 0 || c.Y != 2 {
		t.Errorf("got %+v, want {x:0 Y:2}", c)
	}
}

func Test_sliceReplacesDefault(t *testing.T) {
	t.Parallel()
	c := struct {
		S []string
		T []string
	}{S: []string{"a"}, T: []string{"kept"}}
	if err := config.From(writeConfigFile(t, "S=b c\nT=   ")).To(&c); err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if want := []string{"b", "c"}; !reflect.DeepEqual(c.S, want) {
		t.Errorf("got S %v, want %v", c.S, want)
	}
	if want := []string{"kept"}; !reflect.DeepEqual(c.T, want) {
		t.Errorf("got T %v, want %v", c.T, want)
	}
}

func Test_repeatedDecodeDoesNotAccumulateErrors(t *testing.T) {
	t.Parallel()
	var c struct{ P int }
	b := config.From(writeConfigFile(t, "P=x"))
	want := "config: the following fields had errors: [p]"
	for i := range 2 {
		if err := b.To(&c); err == nil || err.Error() != want {
			t.Errorf("call %d: got err %v, want %q", i, err, want)
		}
	}
}

func Test_fromOptionalOnlyIgnoresMissingFiles(t *testing.T) {
	t.Parallel()
	var c struct{ P int }
	if err := config.FromOptional(filepath.Join(t.TempDir(), "missing")).To(&c); err != nil {
		t.Errorf("missing file: unexpected error %v", err)
	}
	dir := t.TempDir() // reading a directory fails with something other than fs.ErrNotExist
	want := "config: the following fields had errors: [file[" + dir + "]]"
	if err := config.FromOptional(dir).To(&c); err == nil || err.Error() != want {
		t.Errorf("directory: got err %v, want %q", err, want)
	}
}
