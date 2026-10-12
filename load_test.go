package config_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/JeremyLoy/config"
)

func Test_Load(t *testing.T) {
	t.Parallel()
	got, err := config.Load[sourcesConfig](config.FromMap(map[string]string{"HOST": "h", "PORT": "1"}))
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if got.Host != "h" || got.Port != 1 {
		t.Errorf("got %+v, want {Host:h Port:1}", got)
	}

	got, err = config.Load[sourcesConfig](config.FromMap(map[string]string{"HOST": "h", "PORT": "x"}))
	if err == nil || err.Error() != "config: the following fields had errors: [port]" {
		t.Errorf("got err %v, want error for port", err)
	}
	if got.Host != "h" {
		t.Errorf("got %+v, want Host h to be set despite the port error", got)
	}
}

func Test_LoadInvalidTarget(t *testing.T) {
	t.Parallel()
	b := config.FromMap(nil)
	if _, err := config.Load[int](b); !errors.Is(err, config.ErrInvalidTarget) {
		t.Errorf("Load[int]: got err %v, want ErrInvalidTarget", err)
	}
	if _, err := config.Load[*sourcesConfig](b); !errors.Is(err, config.ErrInvalidTarget) {
		t.Errorf("Load[*struct]: got err %v, want ErrInvalidTarget", err)
	}
}

func ExampleLoad() {
	type Config struct {
		Port     int
		LogLevel string `config:"LOG_LEVEL"`
	}
	c, err := config.Load[Config](config.FromMap(map[string]string{
		"PORT":      "8080",
		"LOG_LEVEL": "debug",
	}))

	fmt.Println(c.Port, c.LogLevel, err)

	// Output:
	// 8080 debug <nil>
}
