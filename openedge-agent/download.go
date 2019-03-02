package main

import (
	"fmt"
	"os"
	"path"

	"github.com/baidu/openedge/utils"
	"github.com/baidubce/bce-sdk-go/http"
	"github.com/mholt/archiver"
)

func (m *mo) prepare(all []DatasetInfo) error {
	for _, ds := range all {
		_, err := m.download(ds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mo) download(d DatasetInfo) (string, error) {
	datasetDir := path.Join(m.dir, d.Name, d.Version)
	datasetZipFile := path.Join(datasetDir, d.Name+".zip")

	// dataset exists
	if utils.FileExists(datasetZipFile) {
		return datasetDir, nil
	}

	req := new(http.Request)
	req.SetUri(d.URL)
	res, err := http.Execute(req)
	if err != nil {
		return "", err
	}
	body := res.Body()
	defer body.Close()

	err = utils.WriteFile(datasetZipFile, body)
	if err != nil {
		os.RemoveAll(datasetDir)
		return "", err
	}

	datasetMD5, err := utils.CalculateFileMD5(datasetZipFile)
	if err != nil {
		os.RemoveAll(datasetDir)
		return "", err
	}
	if datasetMD5 != d.MD5 {
		os.RemoveAll(datasetDir)
		return "", fmt.Errorf("dateset (%s) downloaded with unexpected MD5", d.Name)
	}

	err = archiver.Zip.Open(datasetZipFile, datasetDir)
	if err != nil {
		os.RemoveAll(datasetDir)
		return "", err
	}
	return datasetDir, nil
}
