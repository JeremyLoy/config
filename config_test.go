package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"
)

func Test_Integration(t *testing.T) {
	// cannot be Parallelized as it manipulates env vars.
	// It must clear env afterwards to avoid bleeding env changes.
	type D struct {
		E bool
		F bool `config:"feRdiNaND"` // case insensitive
	}
	type testConfig struct {
		A int    `config:"     "` // no-op/ignored
		B string `config:"B"`     // no effect
		C []int
		D D `config:"dOg"` // case insensitive for sub-configs
	}

	file, err := ioutil.TempFile("", "testenv")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(file.Name())

	testData := strings.Join([]string{
		"A=1",
		"B=abc",
		"C=4 5 6",
		"DoG__E=true",
		"DoG__FErDINANd=true",
	}, "\n")
	_, err = file.Write([]byte(testData))
	if err != nil {
		t.Fatalf("failed to write test data to temp file: %v", err)
	}
	err = os.Setenv("B", "overridden")
	if err != nil {
		t.Fatalf("failed to override environ: %v", err)
	}

	var got testConfig
	want := testConfig{
		A: 1,
		B: "overridden",
		C: []int{4, 5, 6},
		D: D{
			E: true,
			F: true,
		},
	}

	From(file.Name()).FromEnv().To(&got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Integration: got %+v, want %+v", got, want)
	}
	os.Clearenv()
}

func Test_stringToSlice(t *testing.T) {
	t.Parallel()
	type args struct {
		in, delim string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{in: "", delim: " "},
			want: nil,
		},
		{
			name: "whitespace",
			args: args{in: "      ", delim: " "},
			want: nil,
		},
		{
			name: "values",
			args: args{in: "  a b c def ghi     ", delim: " "},
			want: []string{"a", "b", "c", "def", "ghi"},
		},
		{
			name: "values - comma delim",
			args: args{in: "  a, b, c ,def ,ghi,     ", delim: ","},
			want: []string{"a", "b", "c", "def", "ghi"},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stringToSlice(tt.args.in, tt.args.delim); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertAndSetValue(t *testing.T) {
	t.Parallel()
	type args struct {
		settable interface{}
		s        string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "int",
			args: args{
				settable: new(int),
				s:        "-2",
			},
			want: func() interface{} { v := -2; return &v }(),
		},
		{
			name: "int8",
			args: args{
				settable: new(int8),
				s:        "-2",
			},
			want: func() interface{} { v := int8(-2); return &v }(),
		},
		{
			name: "int16",
			args: args{
				settable: new(int16),
				s:        "-2",
			},
			want: func() interface{} { v := int16(-2); return &v }(),
		},
		{
			name: "int32",
			args: args{
				settable: new(int32),
				s:        "-2",
			},
			want: func() interface{} { v := int32(-2); return &v }(),
		},
		{
			name: "int64",
			args: args{
				settable: new(int64),
				s:        "-2",
			},
			want: func() interface{} { v := int64(-2); return &v }(),
		},
		{
			name: "uint",
			args: args{
				settable: new(uint),
				s:        "2",
			},
			want: func() interface{} { v := uint(2); return &v }(),
		},
		{
			name: "uint8",
			args: args{
				settable: new(uint8),
				s:        "2",
			},
			want: func() interface{} { v := uint8(2); return &v }(),
		},
		{
			name: "uint16",
			args: args{
				settable: new(uint16),
				s:        "2",
			},
			want: func() interface{} { v := uint16(2); return &v }(),
		},
		{
			name: "uint32",
			args: args{
				settable: new(uint32),
				s:        "2",
			},
			want: func() interface{} { v := uint32(2); return &v }(),
		},
		{
			name: "uint64",
			args: args{
				settable: new(uint64),
				s:        "2",
			},
			want: func() interface{} { v := uint64(2); return &v }(),
		},
		{
			name: "float64",
			args: args{
				settable: new(float64),
				s:        "1.1",
			},
			want: func() interface{} { v := float64(1.1); return &v }(),
		},
		{
			name: "float32",
			args: args{
				settable: new(float32),
				s:        "1.1",
			},
			want: func() interface{} { v := float32(1.1); return &v }(),
		},
		{
			name: "bool",
			args: args{
				settable: new(bool),
				s:        "t",
			},
			want: func() interface{} { v := true; return &v }(),
		},
		{
			name: "string",
			args: args{
				settable: new(string),
				s:        "abc",
			},
			want: func() interface{} { v := "abc"; return &v }(),
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			convertAndSetValue(tt.args.settable, tt.args.s)
			if !reflect.DeepEqual(tt.args.settable, tt.want) {
				t.Errorf("convertAndSetValue = %v, want %v", tt.args.settable, tt.want)
			}
		})
	}
}

func Test_convertAndSetSlice(t *testing.T) {
	t.Parallel()
	type args struct {
		slicePtr interface{}
		values   []string
	}
	tests := []struct {
		name string
		args args
		want interface{}
	}{
		{
			name: "string slice",
			args: args{
				slicePtr: new([]string),
				values:   []string{"a", "b", "c", "def"},
			},
			want: func() interface{} { v := []string{"a", "b", "c", "def"}; return &v },
		},
		{
			name: "int slice",
			args: args{
				slicePtr: new([]int),
				values:   []string{"1", "2", "3", "4"},
			},
			want: func() interface{} { v := []int{1, 2, 3, 4}; return &v },
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			convertAndSetSlice(tt.args.slicePtr, tt.args.values)
		})
	}
}

func Test_stringsToMap(t *testing.T) {
	t.Parallel()
	type args struct {
		ss []string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "expected",
			args: args{
				ss: []string{"", "     ", "=", "A=", "B=1"},
			},
			want: map[string]string{
				"b": "1",
			},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := stringsToMap(tt.args.ss); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringsToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getKey(t *testing.T) {
	t.Parallel()
	type args struct {
		t      reflect.StructField
		prefix string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no tag",
			args: args{
				t: reflect.StructField{
					Name: "name",
					Tag:  "",
				},
				prefix: "pre__",
			},
			want: "pre__name",
		},
		{
			name: "no tag - mixed case",
			args: args{
				t: reflect.StructField{
					Name: "nAMe",
					Tag:  "",
				},
				prefix: "pRe__",
			},
			want: "pre__name",
		},
		{
			name: "empty tag",
			args: args{
				t: reflect.StructField{
					Name: "name",
					Tag:  "config:\"\"",
				},
				prefix: "pre__",
			},
			want: "pre__name",
		},
		{
			name: "whitespace tag",
			args: args{
				t: reflect.StructField{
					Name: "name",
					Tag:  "config:\"    \"",
				},
				prefix: "pre__",
			},
			want: "pre__name",
		},
		{
			name: "tag",
			args: args{
				t: reflect.StructField{
					Name: "name",
					Tag:  "config:\"  tag  \"",
				},
				prefix: "pre__",
			},
			want: "pre__tag",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := getKey(tt.args.t, tt.args.prefix); got != tt.want {
				t.Errorf("getKey() = %v, want %v", got, tt.want)
			}
		})
	}
}
