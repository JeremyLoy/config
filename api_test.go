package config_test

import (
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JeremyLoy/config"
)

func Test_Api(t *testing.T) {
	// cannot be Parallelized as it manipulates env vars.
	// It must clear env afterwards to avoid bleeding env changes.
	type D struct {
		E bool
		F bool `config:"feRdiNaND"` // case insensitive
		J *int
	}
	type testConfig struct {
		A int    `config:"     "` // no-op/ignored
		B string `config:"B"`     // no effect
		C []int
		D D `config:"dOg"` // case-insensitive for sub-configs
		G []int
		H uint8
		I string
		K string
		L time.Duration
		M int8
	}

	file, err := os.CreateTemp("", "testenv")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(file.Name())

	eightHours, err := time.ParseDuration("8h")
	if err != nil {
		panic(err)
	}

	testData := strings.Join([]string{
		"A=1",
		"B=abc",
		"C=4 5 6",
		"DoG__E=true",
		"DoG__FErDINANd=true",
		// should NOT log dog_j as it is not provided
		"G=1 y 2", // should log G[1] as it is an incorrect type, but still work with 0 and 2
		"H=-84",   // should log H as it is an incorrect type
		"I=",      // should NOT log I as there is no way to tell if it is missing or deliberately empty
		"L=8h",
		"M=128", // should fail as it is out of bounds for an int8
	}, "\n")
	_, err = file.Write([]byte(testData))
	if err != nil {
		t.Fatalf("failed to write test data to temp file: %v", err)
	}
	err = os.Setenv("B", "overridden")
	if err != nil {
		t.Fatalf("failed to set environ: %v", err)
	}
	err = os.Setenv("C", "") // this should NOT override as it is empty
	if err != nil {
		t.Fatalf("failed to set environ: %v", err)
	}

	got := testConfig{
		K: "hardcoded",
	}
	var sub D
	want := testConfig{
		A: 1,
		B: "overridden",
		C: []int{4, 5, 6},
		D: D{
			E: true,
			F: true,
			J: nil,
		},
		G: []int{1, 2},
		H: 0,
		I: "",
		K: "hardcoded",
		L: eightHours,
		M: 0,
	}
	wantError := "config: the following fields had errors: [file[nonexistfile] g[1] h m]"

	builder := config.From(file.Name()).From("nonexistfile").FromEnv()
	gotErr := builder.To(&got)
	_ = builder.Sub(&sub, "dog")
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Integration: got %+v, want %+v", got, want)
	}
	if !reflect.DeepEqual(sub, want.D) {
		t.Errorf("Integration: got %+v, want %+v", sub, want.D)
	}
	if gotErr == nil {
		t.Errorf("Integration: should have had an error")
	}
	if !reflect.DeepEqual(gotErr.Error(), wantError) {
		t.Errorf("Integration: gotFailedFields %+v, wantFailedFields %+v", gotErr, wantError)
	}
	os.Clearenv()
}

func Test_shouldPanic(t *testing.T) {
	t.Parallel()

	var s struct{}
	var i int

	tests := []struct {
		name      string
		target    interface{}
		wantPanic bool
	}{
		{
			name:      "*struct",
			target:    &s,
			wantPanic: false,
		},
		{
			name:      "struct",
			target:    s,
			wantPanic: true,
		},
		{
			name:      "*int",
			target:    &i,
			wantPanic: true,
		},
		{
			name:      "int",
			target:    i,
			wantPanic: true,
		},
		{
			name:      "nil",
			target:    nil,
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			defer func() {
				if r := recover(); (r == nil) == tt.wantPanic {
					t.Errorf("should have caused a panic")
				}
			}()
			_ = config.FromEnv().To(tt.target)
		})
	}
}
