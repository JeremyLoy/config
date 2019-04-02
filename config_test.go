package config

import (
	"reflect"
	"testing"
)

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
		t.Run(tt.name, func(t *testing.T) {
			convertAndSetValue(tt.args.settable, tt.args.s)
			if !reflect.DeepEqual(tt.args.settable, tt.want) {
				t.Errorf("convertAndSetValue = %v, want %v", tt.args.settable, tt.want)
			}
		})
	}
}

func Test_convertAndSetSlice(t *testing.T) {
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
		t.Run(tt.name, func(t *testing.T) {
			convertAndSetSlice(tt.args.slicePtr, tt.args.values)
		})
	}
}

func Test_stringsToMap(t *testing.T) {
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
				"B": "1",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := stringsToMap(tt.args.ss); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("stringsToMap() = %v, want %v", got, tt.want)
			}
		})
	}
}
