package auth

import (
	"testing"

	"github.com/baidu/openedge/openedge-hub/config"
	"github.com/stretchr/testify/assert"
)

func TestAuth(t *testing.T) {
	p1 := []config.Principal{{
		Username: "test",
		Password: "hahaha",
		Permissions: []config.Permission{
			{Action: "pub", Permits: []string{"+"}},
			{Action: "sub", Permits: []string{"+"}},
		}},
	}
	au := NewAuth(p1)

	assert.Nil(t, au.AuthenticateAccount("", ""))
	assert.Nil(t, au.AuthenticateAccount("test", ""))
	assert.Nil(t, au.AuthenticateAccount("", "hahaha"))
	assert.Nil(t, au.AuthenticateAccount("test", "hahaha1"))
	assert.Nil(t, au.AuthenticateAccount("test1", "hahaha"))
	authorizer := au.AuthenticateAccount("test", "hahaha")
	assert.NotNil(t, authorizer)

	// "+" strategy validate
	assert.Equal(t, false, authorizer.Authorize(Publish, "/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "a"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "#"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/a"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "/a"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/+"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "a"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "#"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/a"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/a"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/+"))

	// "#" strategy validate
	p2 := []config.Principal{{
		Username: "test",
		Password: "hahaha",
		Permissions: []config.Permission{
			{Action: "pub", Permits: []string{"#"}},
			{Action: "sub", Permits: []string{"#"}},
		}},
	}

	au = NewAuth(p2)
	authorizer = au.AuthenticateAccount("test", "hahaha")
	assert.NotNil(t, authorizer)

	// assert.Equal(t, true, authorizer.Authorize(Publish, "+"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "#"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "/"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "//"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "/+"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "/#"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/+"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/#"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/b"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/#"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/+"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/b"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/+/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/+/a"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/+/#"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/a"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/a/b"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/a/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/a/+"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "+/a/#"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "#"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "/"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "//"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "/+"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "/#"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/+"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/#"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/b"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/#"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/+"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/b"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/+/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/+/a"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/+/#"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/a"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/a/b"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/a/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/a/+"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "+/a/#"))

	// "test/+" strategy validate
	p3 := []config.Principal{{
		Username: "test",
		Password: "hahaha",
		Permissions: []config.Permission{
			{Action: "pub", Permits: []string{"test/+"}},
			{Action: "sub", Permits: []string{"test/+"}},
		}},
	}

	au = NewAuth(p3)
	authorizer = au.AuthenticateAccount("test", "hahaha")
	assert.NotNil(t, authorizer)

	assert.Equal(t, false, authorizer.Authorize(Publish, "test"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "/a"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/+"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/+"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "//"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "//a"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "//+"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "/test/a"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "/test/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/test/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/test/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/test/a"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/test/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+//"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/+/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/+/a"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/+/+"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/c"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "test"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/a"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/+"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/+"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "//"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "//a"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "//+"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/test/a"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/test/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/test/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/test/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/test/a"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/test/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+//"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/+/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/+/a"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/+/+"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/c"))

	// "a/+/b" strategy validate
	p4 := []config.Principal{{
		Username: "test",
		Password: "hahaha",
		Permissions: []config.Permission{
			{Action: "pub", Permits: []string{"a/+/b"}},
			{Action: "sub", Permits: []string{"a/+/b"}},
		}},
	}

	au = NewAuth(p4)
	authorizer = au.AuthenticateAccount("test", "hahaha")
	assert.NotNil(t, authorizer)

	// assert.Equal(t, true, authorizer.Authorize(Publish, "a/+/b"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "a/c/b"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "a//b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/#/b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+/b/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/c/d/b"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a///b"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "test/+/b"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "test/a/b"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "test//b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "test/#/b"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/c"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/#"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "a/+/b"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "a/c/b"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "a//b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/#/b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+/b/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/c/d/b"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a///b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "test/+/b"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "test/a/b"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "test//b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "test/#/b"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/c"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/#"))

	// "test/#" strategy validate
	p5 := []config.Principal{{
		Username: "test",
		Password: "hahaha",
		Permissions: []config.Permission{
			{Action: "pub", Permits: []string{"test/#"}},
			{Action: "sub", Permits: []string{"test/#"}},
		}},
	}

	au = NewAuth(p5)
	authorizer = au.AuthenticateAccount("test", "hahaha")
	assert.NotNil(t, authorizer)

	// assert.Equal(t, false, authorizer.Authorize(Publish, "+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "#"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "//"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "/#"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "+/#"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/b"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/#"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/+"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/b"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/+/b"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/+/"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/+/+"))
	// assert.Equal(t, true, authorizer.Authorize(Publish, "test/a/+/#"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/#"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+/b"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/+/#"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/+"))
	// assert.Equal(t, false, authorizer.Authorize(Publish, "a/b/#"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "#"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "//"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "/#"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "+/#"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/b"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/#"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/+"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/b"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/+/b"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/+/"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/+/+"))
	// assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/a/+/#"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/#"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+/b"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/+/#"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/+"))
	// assert.Equal(t, false, authorizer.Authorize(Subscribe, "a/b/#"))

	principals := []config.Principal{
		{
			Username: "test",
			Password: "hahaha",
			Permissions: []config.Permission{
				{Action: "pub", Permits: []string{"test", "benchmark", "test"}},
				{Action: "sub", Permits: []string{"test", "benchmark"}},
				{Action: "pub", Permits: []string{"test/中文"}},
				{Action: "sub", Permits: []string{"test/中文", "benchmark"}},
			},
		},
		{
			Username: "temp",
			Password: "lalala",
			Permissions: []config.Permission{
				{Action: "pub", Permits: []string{"test", "benchmark"}},
				{Action: "sub", Permits: []string{"test", "benchmark"}},
			},
		},
		{
			Username: "1",
			Permissions: []config.Permission{
				{Action: "pub", Permits: []string{"test", "benchmark"}},
				{Action: "sub", Permits: []string{"test", "benchmark"}},
			},
		},
	}

	// repeat authorizer process
	for _, principal := range principals {
		switch principal.Username {
		case "test":
			pubPermits, subPermits := getPubSubPermits(principal.Permissions)
			assert.Equal(t, 3, len(pubPermits))
			assert.Equal(t, 3, len(subPermits))
		case "temp":
			pubPermits, subPermits := getPubSubPermits(principal.Permissions)
			assert.Equal(t, 2, len(pubPermits))
			assert.Equal(t, 2, len(subPermits))
		case "":
			pubPermits, subPermits := getPubSubPermits(principal.Permissions)
			assert.Equal(t, 2, len(pubPermits))
			assert.Equal(t, 2, len(subPermits))
		}
	}

	au = NewAuth(principals)
	// config username & password authenticate
	assert.NotNil(t, au.AuthenticateAccount("test", "hahaha"))
	assert.Nil(t, au.AuthenticateAccount("test", "lalala"))
	assert.NotNil(t, au.AuthenticateAccount("temp", "lalala"))
	assert.Nil(t, au.AuthenticateAccount("temp", "hahaha"))
	assert.NotNil(t, au.AuthenticateCert("1"))
	assert.Nil(t, au.AuthenticateCert("2"))

	authorizer = au.AuthenticateAccount("test", "hahaha")
	assert.NotNil(t, authorizer)
	// pub permission authorize
	assert.Equal(t, true, authorizer.Authorize(Publish, "test"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "benchmark"))
	assert.Equal(t, true, authorizer.Authorize(Publish, "test/中文"))
	assert.Equal(t, false, authorizer.Authorize(Publish, "temp"))

	// sub permission authorize
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "benchmark"))
	assert.Equal(t, true, authorizer.Authorize(Subscribe, "test/中文"))
	assert.Equal(t, false, authorizer.Authorize(Subscribe, "temp"))
}

func getPubSubPermits(permissions []config.Permission) ([]string, []string) {
	permissions = duplicatePubSubPermitRemove(permissions)
	pubPermits := make([]string, 0)
	subPermits := make([]string, 0)
	for _, permission := range permissions {
		switch permission.Action {
		case Publish:
			pubPermits = permission.Permits
		case Subscribe:
			subPermits = permission.Permits
		}
	}
	return pubPermits, subPermits
}
