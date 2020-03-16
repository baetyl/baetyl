package utils

import "github.com/baetyl/baetyl-core/common"

func MakeKey(resType common.Resource, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(resType) + "/" + name + "/" + ver
}