package sync

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	gosync "sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/mock"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/stretchr/testify/assert"

	"github.com/baetyl/baetyl/v2/node"
	"github.com/baetyl/baetyl/v2/store"
)

func TestObject_FilteringConfig(t *testing.T) {
	cfg := &specv1.Configuration{
		Name:      "ddd",
		Namespace: "sss",
		Data: map[string]string{
			"abc":                                 "cdascd",
			"_object_efg":                         "cdacds",
			"_object_" + context.PlatformString(): "scdasv",
		},
	}
	FilterConfig(cfg)
	assert.Len(t, cfg.Data, 3)
	FilterConfig(cfg)
	assert.Len(t, cfg.Data, 3)
	cfg.Labels = map[string]string{"baetyl-config-type": "baetyl-program"}
	FilterConfig(cfg)
	assert.Len(t, cfg.Data, 2)
	assert.Equal(t, "cdascd", cfg.Data["abc"])
	assert.Equal(t, "scdasv", cfg.Data["_object_"+context.PlatformString()])
	FilterConfig(cfg)
	assert.Len(t, cfg.Data, 2)
	assert.Equal(t, "cdascd", cfg.Data["abc"])
}

func TestObject_DownloadObject(t *testing.T) {
	f, err := os.CreateTemp("", t.Name())
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
	dir := t.TempDir()
	fmt.Println("-->tempdir", dir)
	file1 := filepath.Join(dir, "file1")
	os.WriteFile(file1, content, 0644)

	var responses []*mock.Response
	for i := 0; i < 34; i++ {
		responses = append(responses, mock.NewResponse(200, content))
	}
	assert.NoError(t, err)
	objMs := mock.NewServer(nil, responses...)
	assert.NotNil(t, objMs)
	defer objMs.Close()

	cc := http.ClientConfig{}
	err = utils.UnmarshalYAML(nil, &cc)
	assert.NoError(t, err)
	cc.Address = objMs.URL
	cc.CA = "./testcert/ca.pem"
	cc.Key = "./testcert/client.key"
	cc.Cert = "./testcert/client.pem"
	cc.InsecureSkipVerify = true

	//syn, err := NewSync(sc, sto, nod)
	ops, err := cc.ToClientOptions()
	assert.NoError(t, err)
	cli := http.NewClient(ops)

	md5, _ := utils.CalculateFileMD5(file1)
	obj := &specv1.ConfigurationObject{
		URL: objMs.URL,
		MD5: md5,
	}
	// already exist
	err = downloadObject(cli, obj, dir, file1, "")
	assert.NoError(t, err)

	// normal download
	file2 := filepath.Join(dir, "file2")
	err = downloadObject(cli, obj, dir, file2, "")
	assert.NoError(t, err)

	// invalid url
	file3 := filepath.Join(dir, "invalidUrl")
	obj.URL = "http:xxx"
	err = downloadObject(cli, obj, dir, file3, "")
	assert.Error(t, err)
	obj.URL = objMs.URL

	// not zip file
	file4 := filepath.Join(dir, "file4")
	obj.MD5 = md5
	err = downloadObject(cli, obj, dir, file4, "zip")
	assert.Error(t, err)

	// download file not exist (multiple routine)
	file5 := filepath.Join(dir, "file5")
	var wg gosync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *gosync.WaitGroup) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := downloadObject(cli, obj, dir, file5, "")
			assert.NoError(t, err)
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	res, err := os.ReadFile(file5)
	assert.NoError(t, err)
	assert.Equal(t, res, content)

	// download file which already exist (multiple routine)
	file6 := filepath.Join(dir, "file6")
	os.WriteFile(file6, content, 0644)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *gosync.WaitGroup) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := downloadObject(cli, obj, dir, file6, "")
			assert.NoError(t, err)
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	res, err = os.ReadFile(file6)
	assert.NoError(t, err)
	assert.Equal(t, res, content)

	// download file with wrong content exist (multiple routine)
	file7 := filepath.Join(dir, "file7")
	os.WriteFile(file7, []byte("wrong"), 0644)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(wg *gosync.WaitGroup) {
			time.Sleep(time.Millisecond * time.Duration(rand.Intn(100)))
			err := downloadObject(cli, obj, dir, file7, "")
			assert.NoError(t, err)
			wg.Done()
		}(&wg)
	}
	wg.Wait()
	res, err = os.ReadFile(file7)
	assert.NoError(t, err)
	assert.Equal(t, res, content)
}
