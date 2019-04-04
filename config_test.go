package config

import (
	"io/ioutil"
	"os"
	"reflect"
	"testing"
)

func Test_Integration(t *testing.T) {
	t.Parallel()
	type D struct {
		E bool
	}
	type testConfig struct {
		A int
		B string
		C []int
		D D
	}

	file, err := ioutil.TempFile("", "testenv")
	if err != nil {
		t.Fatalf("failed to create temporary file: %v", err)
	}
	defer os.Remove(file.Name())

	testData := "A=1\nB=abc\nC=4 5 6\nD__E=true"
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
		},
	}

	From(file.Name()).FromEnv().To(&got)
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Integration: got %+v, want %+v", got, want)
	}

}

func Test_stringToSlice(t *testing.T) {
	type args struct {
		in string
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "empty",
			args: args{in: ""},
			want: nil,
		},
		{
			name: "whitespace",
			args: args{in: "      "},
			want: nil,
		},
		{
			name: "values",
			args: args{in: "  a b c def ghi     "},
			want: []string{"a", "b", "c", "def", "ghi"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringToSlice(tt.args.in); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringToSlice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_convertAndSetValue(t *testing.T) {
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

func Test_filesystemErrors(t *testing.T) {
	t.Parallel()

	var c struct{}
	if err := From("a non existantFile").To(&c).Err(); err == nil {
		t.Error("expected an error but got none.")
	}
}

func Test_conversionErrors(t *testing.T) {
	tests := []struct {
		name     string
		settable interface{}
	}{
		{
			name:     "int",
			settable: new(int),
		},
		{
			name:     "int8",
			settable: new(int8),
		},
		{
			name:     "int16",
			settable: new(int16),
		},
		{
			name:     "int32",
			settable: new(int32),
		},
		{
			name:     "int64",
			settable: new(int64),
		},
		{
			name:     "uint",
			settable: new(uint),
		},
		{
			name:     "uint8",
			settable: new(uint8),
		},
		{
			name:     "uint16",
			settable: new(uint16),
		},
		{
			name:     "uint32",
			settable: new(uint32),
		},
		{
			name:     "uint64",
			settable: new(uint64),
		},
		{
			name:     "float64",
			settable: new(float64),
		},
		{
			name:     "float32",
			settable: new(float32),
		},
		{
			name:     "bool",
			settable: new(bool),
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := convertAndSetValue(tt.settable, "!"); err == nil {
				t.Errorf("expected an error from convertAndSetValue")
			}
		})
	}
}
