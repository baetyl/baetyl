package sync

import (
	"fmt"
	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-core/node"
	"github.com/baetyl/baetyl-core/store"
	"github.com/baetyl/baetyl-go/mock"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/stretchr/testify/assert"
)

func TestSyncDownloadObject(t *testing.T) {
	f, err := ioutil.TempFile("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, f)
	fmt.Println("-->tempfile", f.Name())

	sto, err := store.NewBoltHold(f.Name())
	assert.NoError(t, err)
	assert.NotNil(t, sto)

	nod, err := node.NewNode(sto)
	assert.NoError(t, err)
	assert.NotNil(t, nod)

	content := []byte("test")
	dir, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	assert.NotNil(t, dir)
	fmt.Println("-->tempdir", dir)
	file1 := filepath.Join(dir, "file1")
	ioutil.WriteFile(file1, content, 0644)

	assert.NoError(t, err)
	objMs := mock.NewServer(nil,
		mock.NewResponse(200, content), mock.NewResponse(200, content),
		mock.NewResponse(200, content), mock.NewResponse(200, content))
	assert.NotNil(t, objMs)
	defer objMs.Close()

	sc := config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Cloud.HTTP.Address = objMs.URL
	sc.Cloud.HTTP.CA = "./testcert/ca.pem"
	sc.Cloud.HTTP.Key = "./testcert/client.key"
	sc.Cloud.HTTP.Cert = "./testcert/client.pem"
	sc.Cloud.HTTP.InsecureSkipVerify = true

	//syn, err := NewSync(sc, sto, nod)
	ops, err := sc.Cloud.HTTP.ToClientOptions()
	assert.NoError(t, err)
	syn := &sync{
		store: sto,
		nod:   nod,
		http:  http.NewClient(ops),
		log:   log.With(log.Any("test", "sync")),
	}

	md5, _ := utils.CalculateFileMD5(file1)
	obj := &specv1.ConfigurationObject{
		URL: objMs.URL,
		MD5: md5,
	}
	// already exist
	err = syn.downloadObject(obj, dir, file1, false)
	assert.NoError(t, err)

	// normal download
	file2 := filepath.Join(dir, "file2")
	err = syn.downloadObject(obj, dir, file2, false)
	assert.NoError(t, err)

	// invalid url
	file3 := filepath.Join(dir, "invalidUrl")
	obj.URL = ""
	err = syn.downloadObject(obj, dir, file3, false)
	assert.Error(t, err)

	// failed to write into file
	os.Chmod(dir, 0444)
	file4 := filepath.Join(dir, "file4")
	obj.URL = objMs.URL
	err = syn.downloadObject(obj, dir, file4, false)
	assert.Error(t, err)
	os.Chmod(dir, 0755)

	// md5 error
	file5 := filepath.Join(dir, "file5")
	obj.MD5 = ""
	err = syn.downloadObject(obj, dir, file5, false)
	assert.NoError(t, err)

	// not zip file
	file6 := filepath.Join(dir, "file6")
	obj.MD5 = md5
	err = syn.downloadObject(obj, dir, file6, true)
	assert.Error(t, err)
}
