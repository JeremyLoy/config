package config_test

import (
	"errors"
	"fmt"
	"io/fs"
	"reflect"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/JeremyLoy/config"
)

type sourcesConfig struct {
	Host  string
	Port  int
	Hosts []string
}

func Test_sources(t *testing.T) {
	t.Parallel()
	const content = "# a comment\n  # an indented comment\n\nHOST=example.com\n#PORT=1\nPORT=8080\n"
	want := sourcesConfig{Host: "example.com", Port: 8080}

	fsys := fstest.MapFS{"app.env": {Data: []byte(content)}}
	tests := []struct {
		name    string
		builder *config.Builder
	}{
		{name: "From", builder: config.From(writeConfigFile(t, content))},
		{name: "FromFS", builder: config.FromFS(fsys, "app.env")},
		{name: "FromReader", builder: config.FromReader(strings.NewReader(content))},
		{name: "FromMap", builder: config.FromMap(map[string]string{"host": "example.com", "PoRt": "8080", "HOSTS": ""})},
		{name: "chained", builder: config.FromMap(map[string]string{"HOST": "overridden", "PORT": "1"}).FromReader(strings.NewReader(content))},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var got sourcesConfig
			if err := tt.builder.To(&got); err != nil {
				t.Fatalf("unexpected error %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("got %+v, want %+v", got, want)
			}
		})
	}
}

func Test_fromFSMissingFile(t *testing.T) {
	t.Parallel()
	var c sourcesConfig
	err := config.FromFS(fstest.MapFS{}, "missing.env").To(&c)
	if !errors.Is(err, fs.ErrNotExist) || err.Error() != "config: the following fields had errors: [file[missing.env]]" {
		t.Errorf("got err %v, want fs.ErrNotExist for file[missing.env]", err)
	}
}

func Example_fromMap() {
	var c sourcesConfig
	config.FromMap(map[string]string{
		"HOST":  "localhost",
		"PORT":  "8080",
		"HOSTS": "a b c",
	}).To(&c)

	fmt.Println(c.Host, c.Port, c.Hosts)

	// Output:
	// localhost 8080 [a b c]
}
