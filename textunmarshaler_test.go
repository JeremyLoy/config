package config_test

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/netip"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JeremyLoy/config"
)

// upper is a custom encoding.TextUnmarshaler that rejects empty-after-trim input.
type upper string

func (u *upper) UnmarshalText(b []byte) error {
	if strings.TrimSpace(string(b)) == "" || strings.Contains(string(b), "!") {
		return fmt.Errorf("bad upper %q", b)
	}
	*u = upper(strings.ToUpper(string(b)))
	return nil
}

func Test_textUnmarshaler(t *testing.T) {
	t.Parallel()
	type Sub struct {
		Started time.Time
	}
	var c struct {
		Addr    netip.Addr
		Addrs   []netip.Addr
		IP      net.IP // a []byte, but must be parsed as a single value
		Level   slog.Level
		Name    upper
		Bad     upper
		Sub     Sub
		Started time.Time
	}
	file := writeConfigFile(t, strings.Join([]string{
		"ADDR=10.0.0.1",
		"ADDRS=::1 127.0.0.1",
		"IP=192.168.0.1",
		"LEVEL=warn",
		"NAME=gopher",
		"BAD=secret!",
		"SUB__STARTED=2026-08-01T00:00:00Z",
		"STARTED=2026-02-01T00:00:00Z",
	}, "\n"))

	err := config.From(file).To(&c)

	if err == nil || err.Error() != "config: the following fields had errors: [bad]" {
		t.Fatalf("got err %v, want error for bad", err)
	}
	if !errors.Is(err, config.ErrInvalidValue) || strings.Contains(err.(*config.Error).Errors[0].Error(), "secret") {
		t.Errorf("got %v, want sanitized ErrInvalidValue", err.(*config.Error).Errors[0])
	}
	if want := netip.MustParseAddr("10.0.0.1"); c.Addr != want {
		t.Errorf("Addr = %v, want %v", c.Addr, want)
	}
	if want := []netip.Addr{netip.MustParseAddr("::1"), netip.MustParseAddr("127.0.0.1")}; !reflect.DeepEqual(c.Addrs, want) {
		t.Errorf("Addrs = %v, want %v", c.Addrs, want)
	}
	if want := net.ParseIP("192.168.0.1"); !c.IP.Equal(want) {
		t.Errorf("IP = %v, want %v", c.IP, want)
	}
	if c.Level != slog.LevelWarn {
		t.Errorf("Level = %v, want %v", c.Level, slog.LevelWarn)
	}
	if c.Name != "GOPHER" {
		t.Errorf("Name = %q, want GOPHER", c.Name)
	}
	if want := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC); !c.Sub.Started.Equal(want) {
		t.Errorf("Sub.Started = %v, want %v", c.Sub.Started, want)
	}
	if want := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC); !c.Started.Equal(want) {
		t.Errorf("Started = %v, want %v", c.Started, want)
	}
}

func Example_textUnmarshaler() {
	os.Clearenv()
	os.Setenv("LISTEN", "127.0.0.1")
	os.Setenv("LOG_LEVEL", "debug")

	// Any type implementing encoding.TextUnmarshaler is supported.
	var c struct {
		Listen   netip.Addr
		LogLevel slog.Level `config:"LOG_LEVEL"`
	}
	config.FromEnv().To(&c)

	fmt.Println(c.Listen, c.LogLevel)

	// Output:
	// 127.0.0.1 DEBUG
}
