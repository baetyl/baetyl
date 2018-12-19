package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestEqual(t *testing.T) {
	aI := []byte{1, 2}
	bI := []byte{1, 2}
	bI2 := []byte{1}
	fmt.Printf("%p %p\n", &aI, &bI)
	type args struct {
		a interface{}
		b interface{}
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "struct-1",
			args: args{
				a: struct {
					I int
					S string
					F float64
				}{I: 1, S: "s", F: 0.5},
				b: struct {
					I int
					S string
					F float64
				}{I: 1, S: "s", F: .50},
			},
			want: true,
		},
		{
			name: "struct-2",
			args: args{
				a: struct {
					I int
					S string
					F float64
				}{I: 1, S: "s", F: 0.5},
				b: struct {
					I int
					S string
					F float64
				}{I: 1, S: "s", F: 1.50},
			},
			want: false,
		},
		{
			name: "struct-3",
			args: args{
				a: struct {
					I []byte
				}{I: aI},
				b: struct {
					I []byte
				}{I: bI},
			},
			want: true,
		},
		{
			name: "struct-4",
			args: args{
				a: struct {
					I []byte
				}{I: aI},
				b: struct {
					I []byte
				}{I: bI2},
			},
			want: false,
		},
		{
			name: "struct-5",
			args: args{
				a: struct {
					I *[]byte
				}{I: &aI},
				b: struct {
					I *[]byte
				}{I: &bI},
			},
			want: true,
		},
		{
			name: "struct-5",
			args: args{
				a: struct {
					I time.Duration
				}{I: time.Second * 60},
				b: struct {
					I time.Duration
				}{I: time.Minute},
			},
			want: true,
		},
		{
			name: "struct-6",
			args: args{
				a: struct {
					I time.Duration
				}{I: time.Second},
				b: struct {
					I time.Duration
				}{I: time.Minute},
			},
			want: false,
		},
		{
			name: "struct-7",
			args: args{
				a: struct {
					inner *args
				}{inner: &args{a: "a", b: "b"}},
				b: struct {
					inner *args
				}{inner: &args{a: "a", b: "b"}},
			},
			want: true,
		},
		{
			name: "struct-7",
			args: args{
				a: struct {
					inner *args
				}{inner: &args{a: "a", b: "b"}},
				b: struct {
					inner *args
				}{inner: &args{a: "a"}},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Equal(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Equal() = %v, want %v", got, tt.want)
			}
		})
	}
}
