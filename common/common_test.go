package common

import (
	"reflect"
	"testing"
)

func TestStringToBytes(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want []byte
	}{
		{
			name: "Basic string",
			args: args{s: "hello"},
			want: []byte("hello"),
		},
		{
			name: "Empty string",
			args: args{s: ""},
			want: nil,
		},
		{
			name: "String with special characters",
			args: args{s: "!@#$%^&*()123"},
			want: []byte("!@#$%^&*()123"),
		},
		{
			name: "String with Unicode characters",
			args: args{s: "你好, world"},
			want: []byte("你好, world"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StringToBytes(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("StringToBytes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBytesToString(t *testing.T) {
	type args struct {
		b []byte
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Basic byte slice",
			args: args{b: []byte("hello")},
			want: "hello",
		},
		{
			name: "Nil byte slice",
			args: args{b: nil},
			want: "",
		},
		{
			name: "Empty byte slice",
			args: args{b: []byte{}},
			want: "",
		},
		{
			name: "Byte slice with special characters",
			args: args{b: []byte("!@#$%^&*()123")},
			want: "!@#$%^&*()123",
		},
		{
			name: "Byte slice with Unicode characters",
			args: args{b: []byte("你好, world")},
			want: "你好, world",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := BytesToString(tt.args.b); got != tt.want {
				t.Errorf("BytesToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
