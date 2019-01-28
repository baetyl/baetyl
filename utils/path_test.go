package utils

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
}

// func TestParseVolumes(t *testing.T) {
// 	valumes := []string{
// 		"var/db/openedge/module/m1:/m1/config:ro",
// 		"var/db/openedge/volume/m1:/m1/data",
// 		"var/log/openedge/m1:/m1/log",
// 		"/m1/",
// 	}
// 	result, err := ParseVolumes("/work/dir", valumes)
// 	assert.NoError(t, err)
// 	assert.EqualValues(t, []string{
// 		"/work/dir/var/db/openedge/module/m1:/m1/config:ro",
// 		"/work/dir/var/db/openedge/volume/m1:/m1/data",
// 		"/work/dir/var/log/openedge/m1:/m1/log",
// 		"/work/dir/m1:/m1",
// 	}, result)

// 	valumes = []string{
// 		"../../../:/m1",
// 	}
// 	result, err = ParseVolumes("/work/dir", valumes)
// 	assert.EqualError(t, err, "volume (../../../:/m1) contains invalid string (..)")
// 	assert.Nil(t, result)
// 	valumes = []string{
// 		"m1",
// 	}
// 	result, err = ParseVolumes("/work/dir", valumes)
// 	assert.EqualError(t, err, "volume (m1) in container is not absolute")
// 	assert.Nil(t, result)
// }
