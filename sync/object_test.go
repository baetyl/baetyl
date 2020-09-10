package sync

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"path/filepath"
	gosync "sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/mock"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
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

	var responses []*mock.Response
	for i := 0; i < 34; i++ {
		responses = append(responses, mock.NewResponse(200, content))
	}
	assert.NoError(t, err)
	objMs := mock.NewServer(nil, responses...)
	assert.NotNil(t, objMs)
	defer objMs.Close()

	sc := config.SyncConfig{}
	err = utils.UnmarshalYAML(nil, &sc)
	assert.NoError(t, err)
	sc.Download.Address = objMs.URL
	sc.Download.CA = "./testcert/ca.pem"
	sc.Download.Key = "./testcert/client.key"
	sc.Download.Cert = "./testcert/client.pem"
	sc.Download.InsecureSkipVerify = true

	//syn, err := NewSync(sc, sto, nod)
	ops, err := sc.Download.ToClientOptions()
	assert.NoError(t, err)
	syn := &sync{
		store:    sto,
		nod:      nod,
		download: http.NewClient(ops),
		log:      log.With(log.Any("test", "sync")),
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
	obj.URL = objMs.URL

	// not zip file
	file4 := filepath.Join(dir, "file4")
	obj.MD5 = md5
	err = syn.downloadObject(obj, dir, file4, true)
	assert.Error(t, err)

	// download file not exist (multiple routine)
	file5 := filepath.Join(dir, "file5")
	var wg gosync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *gosync.WaitGroup) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := syn.downloadObject(obj, dir, file5, false)
			assert.NoError(t, err)
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	res, err := ioutil.ReadFile(file5)
	assert.NoError(t, err)
	assert.Equal(t, res, content)

	// download file which already exist (multiple routine)
	file6 := filepath.Join(dir, "file6")
	ioutil.WriteFile(file6, content, 0644)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *gosync.WaitGroup) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := syn.downloadObject(obj, dir, file6, false)
			assert.NoError(t, err)
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	res, err = ioutil.ReadFile(file6)
	assert.NoError(t, err)
	assert.Equal(t, res, content)

	// download file with wrong content exist (multiple routine)
	file7 := filepath.Join(dir, "file7")
	ioutil.WriteFile(file7, []byte("wrong"), 0644)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *gosync.WaitGroup) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := syn.downloadObject(obj, dir, file7, false)
			assert.NoError(t, err)
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	res, err = ioutil.ReadFile(file7)
	assert.NoError(t, err)
	assert.Equal(t, res, content)
}
