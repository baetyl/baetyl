package utils

import (
	"fmt"
	"net/url"
	"strings"
)

// ParseURL parses a url string
func ParseURL(addr string) (*url.URL, error) {
	parts := strings.SplitN(addr, "://", 2)
	if len(parts) == 1 {
		return nil, fmt.Errorf("failed to parse address (%s)", addr)
	}

	var basePath string
	proto, addr := parts[0], parts[1]
	if proto == "tcp" {
		parsed, err := url.Parse("tcp://" + addr)
		if err != nil {
			return nil, err
		}
		addr = parsed.Host
		basePath = parsed.Path
	}
	return &url.URL{
		Scheme: proto,
		Host:   addr,
		Path:   basePath,
	}, nil
}
