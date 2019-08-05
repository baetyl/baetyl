package utils

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathExists(t *testing.T) {
	dirpath, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dirpath)

	assert.True(t, PathExists(dirpath))
	assert.False(t, PathExists(path.Join(dirpath, "nonexist")))

	filepath, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(filepath.Name())

	assert.True(t, PathExists(filepath.Name()))
	assert.False(t, PathExists(filepath.Name()+"-nonexist"))
}

func TestDirExists(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	assert.True(t, DirExists(dir))
	assert.False(t, DirExists(path.Join(dir, "nonexist")))

	file, err := ioutil.TempFile(dir, "")
	assert.False(t, DirExists(file.Name()))
}

func TestFileExists(t *testing.T) {
	file, err := ioutil.TempFile("", "")
	assert.NoError(t, err)
	defer os.RemoveAll(file.Name())

	assert.True(t, FileExists(file.Name()))
	assert.False(t, FileExists(path.Dir(file.Name())))
	assert.False(t, FileExists(file.Name()+"-nonexist"))
}

func TestPathJoin(t *testing.T) {
	assert.Equal(t, "/var/db/a", path.Join("/var/db", "a"))
	assert.Equal(t, "/var/db/var", path.Join("/var/db", "/var"))
	assert.Equal(t, "/var/db/var/db/a", path.Join("/var/db", "/var/db/a"))
	assert.Equal(t, "/var/db/var/db/a", path.Join("/var/db", "var/db/a"))
	assert.Equal(t, "/var/db/a/b", path.Join("/var/db", "a/b/c/./.."))
	p, err := filepath.Rel("var/db/openedge", "var/db/openedge/vv/v1")
	assert.NoError(t, err)
	assert.Equal(t, "vv/v1", p)
	p, err = filepath.Rel("var/db/openedge", "var/db/openedge/../../../vv/v1")
	assert.NoError(t, err)
	assert.Equal(t, "../../../vv/v1", p)
	assert.Equal(t, "../../../vv/v1", path.Clean(p))
	assert.Equal(t, "vv/v1", path.Join("var/db/openedge", p))
	assert.False(t, path.IsAbs(p))
	assert.False(t, path.IsAbs("var/db/openedge/./vv/v1"))
	assert.False(t, path.IsAbs("var/db/openedge/vv/v1"))
}

func TestEscapeIntercept(t *testing.T) {
	tests := []struct {
		args string
		want string
	}{
		{
			args: "bin/../../../test/edge/bin/../op",
			want: "test/edge/op",
		},
		{
			args: "bin/../../yiu..//.../.././",
			want: "yiu..",
		},
		{
			args: "bin/../../",
			want: ".",
		},
		{
			args: "./",
			want: ".",
		},
		{
			args: "./bin/openedge/../var/db/test",
			want: "bin/var/db/test",
		},
		{
			args: "./bin/openedge/var/db/test",
			want: "bin/openedge/var/db/test",
		},
		{
			args: "bin/openedge/var/db/test",
			want: "bin/openedge/var/db/test",
		},
		{
			args: "/etc/",
			want: "etc",
		},
		{
			args: "/bin/../../..//openedge/var/db/test",
			want: "openedge/var/db/test",
		},
		{
			args: "/bin/../../../",
			want: ".",
		},
		{
			args: "/",
			want: ".",
		},
		{
			args: "",
			want: ".",
		},
	}
	for _, test := range tests {
		str, err := EscapeIntercept(test.args)
		assert.NoError(t, err)
		assert.Equal(t, test.want, str)
	}
}
