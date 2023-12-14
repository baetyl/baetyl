// Package roam 大文件同步模块，用于集群环境部署下大文件在集群中的传播
package roam

import (
	"io"
	"mime/multipart"
	gohttp "net/http"
	"net/url"
	"os"
	"sync"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
)

const (
	HeaderKeyAccept      = "accept"
	HeaderKeyContentType = "content-type"
)

var (
	ErrRoamUpload = errors.New("failed to upload file")
)

type fileRoamer struct {
	dir    string
	file   string
	md5    string
	unpack string
	url    url.URL
	cli    *http.Client
	errs   *[]string
	wg     *sync.WaitGroup
	mx     *sync.Mutex
	log    *log.Logger
}

func (fr *fileRoamer) Uploading() error {
	fr.log.Debug("roam uploading start")
	r, w := io.Pipe()
	defer func() {
		r.Close()
		fr.wg.Done()
		fr.log.Debug("roam upload file complete")
	}()

	m := multipart.NewWriter(w)
	// reading文件读取失败会导致reader异常关闭，使Post动作出错，在post处会捕获错误写入errs并退出
	// 这里采用pipe读取文件是为了防止直接打开读入导致mem使用过大
	go reading(fr.file, m, w)

	header := map[string]string{
		HeaderKeyAccept:              "*/*",
		HeaderKeyContentType:         m.FormDataContentType(),
		specv1.HeaderKeyObjectDir:    fr.dir,
		specv1.HeaderKeyObjectMD5:    fr.md5,
		specv1.HeaderKeyObjectUnpack: fr.unpack,
	}

	fr.log.Debug("roam uploading", log.Any("url", fr.url.String()), log.Any("header", header))
	resp, err := fr.cli.PostURL(fr.url.String(), r, header)
	if err != nil {
		fr.log.Debug("roam upload file error", log.Error(err), log.Any("url", fr.url.String()))
		fr.mx.Lock()
		if fr.errs != nil {
			*fr.errs = append(*fr.errs, err.Error())
		}
		fr.mx.Unlock()
		return errors.Trace(err)
	}
	if resp.StatusCode == gohttp.StatusOK {
		fr.log.Info("upload file success", log.Any("file", fr.file),
			log.Any("url", fr.url.String()), log.Any("response", resp.StatusCode))
	} else {
		fr.log.Error(ErrRoamUpload.Error(), log.Any("file", fr.file),
			log.Any("url", fr.url.String()), log.Any("response", resp.StatusCode))
		fr.mx.Lock()
		if fr.errs != nil {
			*fr.errs = append(*fr.errs, ErrRoamUpload.Error())
		}
		fr.mx.Unlock()
		return errors.Trace(ErrRoamUpload)
	}

	return nil
}

func reading(f string, writer *multipart.Writer, w *io.PipeWriter) error {
	defer func() {
		writer.Close()
		w.Close()
		log.L().Debug("roam reading close")
	}()

	part, err := writer.CreateFormFile("", f)
	if err != nil {
		return errors.Trace(err)
	}
	file, err := os.Open(f)
	if err != nil {
		return errors.Trace(err)
	}
	defer file.Close()

	if _, err = io.Copy(part, file); err != nil {
		return errors.Trace(err)
	}
	return nil
}
