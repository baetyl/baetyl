package utils

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParseURL(t *testing.T) {
	tests := []struct {
		name    string
		addr    string
		want    *url.URL
		wantErr bool
	}{
		{
			name: "unix-1",
			addr: "unix:///var/run/baetyl.sock",
			want: &url.URL{
				Scheme: "unix",
				Host:   "/var/run/baetyl.sock",
			},
		},
		{
			name: "unix-2",
			addr: "unix://./var/run/baetyl.sock",
			want: &url.URL{
				Scheme: "unix",
				Host:   "./var/run/baetyl.sock",
			},
		},
		{
			name: "unix-3",
			addr: "unix://var/run/baetyl.sock",
			want: &url.URL{
				Scheme: "unix",
				Host:   "var/run/baetyl.sock",
			},
		},
		{
			name: "tcp-1",
			addr: "tcp://127.0.0.1:50050",
			want: &url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:50050",
			},
		},
		{
			name: "tcp-2",
			addr: "tcp://127.0.0.1:50050/v1/api",
			want: &url.URL{
				Scheme: "tcp",
				Host:   "127.0.0.1:50050",
				Path:   "/v1/api",
			},
		},
		{
			name: "http-1",
			addr: "http://127.0.0.1:50050/v1/api",
			want: &url.URL{
				Scheme: "http",
				Host:   "127.0.0.1:50050",
				Path:   "/v1/api",
			},
		},
		{
			name: "https-1",
			addr: "https://127.0.0.1:50050/v1/api",
			want: &url.URL{
				Scheme: "https",
				Host:   "127.0.0.1:50050",
				Path:   "/v1/api",
			},
		},
		{
			name: "error-1",
			want: &url.URL{},
		},
		{
			name: "error-2",
			addr: "unix://",
			want: &url.URL{
				Scheme: "unix",
			},
		},
		{
			name: "error-3",
			addr: "dummy",
			want: &url.URL{
				Path: "dummy",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseURL(tt.addr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
