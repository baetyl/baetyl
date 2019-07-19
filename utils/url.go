package utils

import (
	"net/url"
	"strings"
)

// ParseURL parses a url string
func ParseURL(addr string) (*url.URL, error) {
	if strings.HasPrefix(addr, "unix://") {
		parts := strings.SplitN(addr, "://", 2)
		return &url.URL{
			Scheme: parts[0],
			Host:   parts[1],
		}, nil
	}
	return url.Parse(addr)
}
