package config

import (
	"reflect"
	"testing"
)

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
		name    string
		args    args
		want    interface{}
		wantErr bool
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
		{
			name: "bad convert",
			args: args{
				settable: new(int),
				s:        "abc",
			},
			want:    func() interface{} { v := 0; return &v }(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ok := convertAndSetValue(reflect.ValueOf(tt.args.settable), tt.args.s)
			if !reflect.DeepEqual(tt.args.settable, tt.want) {
				t.Errorf("convertAndSetValue = %v, want %v", tt.args.settable, tt.want)
			}
			if ok == tt.wantErr {
				t.Errorf("convertAndSetValue ok = %v, wantErr %v", ok, tt.wantErr)
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
		name    string
		args    args
		want    interface{}
		wantErr []int
	}{
		{
			name: "string slice",
			args: args{
				slicePtr: new([]string),
				values:   []string{"a", "b", "c", "def"},
			},
			want: func() interface{} { v := []string{"a", "b", "c", "def"}; return &v }(),
		},
		{
			name: "int slice",
			args: args{
				slicePtr: new([]int),
				values:   []string{"1", "2", "3", "4"},
			},
			want: func() interface{} { v := []int{1, 2, 3, 4}; return &v }(),
		},
		{
			name: "int slice - bad values",
			args: args{
				slicePtr: new([]int),
				values:   []string{"1", "x", "3", "4"},
			},
			want:    func() interface{} { v := []int{1, 3, 4}; return &v }(),
			wantErr: []int{1},
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotErr := convertAndSetSlice(reflect.ValueOf(tt.args.slicePtr), tt.args.values)
			if !reflect.DeepEqual(tt.args.slicePtr, tt.want) {
				t.Errorf("convertAndSetSlice = %v, want: %v", tt.args.slicePtr, tt.want)
			}
			if !reflect.DeepEqual(gotErr, tt.wantErr) {
				t.Errorf("convertAndSetSlice err = %v, want: %v", gotErr, tt.wantErr)
			}

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
