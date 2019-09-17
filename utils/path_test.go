package utils

import (
	"io"
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
	p, err := filepath.Rel("var/db/baetyl", "var/db/baetyl/vv/v1")
	assert.NoError(t, err)
	assert.Equal(t, "vv/v1", p)
	p, err = filepath.Rel("var/db/baetyl", "var/db/baetyl/../../../vv/v1")
	assert.NoError(t, err)
	assert.Equal(t, "../../../vv/v1", p)
	assert.Equal(t, "../../../vv/v1", path.Clean(p))
	assert.Equal(t, "vv/v1", path.Join("var/db/baetyl", p))
	assert.False(t, path.IsAbs(p))
	assert.False(t, path.IsAbs("var/db/baetyl/./vv/v1"))
	assert.False(t, path.IsAbs("var/db/baetyl/vv/v1"))
	assert.Equal(t, "/usr/local/bin", path.Join("/usr/local/", path.Join("/", "../../../../bin")))
	assert.Equal(t, "/mnt/data0", path.Join("/", path.Join("/", "/mnt/data0")))
}

func TestCreateSymlink(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	assert.NoError(t, err)
	cwd, err := os.Getwd()
	assert.NoError(t, err)
	os.Chdir(dir)
	filename := "file"
	f, err := os.Create(filename)
	assert.NoError(t, err)
	defer f.Close()
	content := "test"
	_, err = io.WriteString(f, content)
	assert.NoError(t, err)
	symlink := "symlink"
	CreateSymlink(filename, symlink)
	res, err := ioutil.ReadFile(symlink)
	assert.NoError(t, err)
	assert.Equal(t, content, string(res))
	os.Chdir(cwd)
}
