package utils

import "regexp"

// IsClientID checks clientID
func IsClientID(v string) bool {
	if v == "" {
		return true
	}
	r := regexp.MustCompile("^[0-9A-Za-z_-]{1,128}$")
	return r.MatchString(v)
}
