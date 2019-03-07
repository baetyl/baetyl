package main

import (
	"fmt"
	"os"
	"path"

	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	"github.com/baidubce/bce-sdk-go/http"
	"github.com/mholt/archiver"
)

func (m *mo) prepare(all []openedge.VolumeInfo) error {
	for _, ds := range all {
		if ds.Meta.URL == "" {
			continue
		}
		_, err := m.download(ds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mo) download(v openedge.VolumeInfo) (string, error) {
	volumeDir := path.Join(m.dir, path.Clean(v.Path))
	volumeZipFile := path.Join(volumeDir, v.Name+".zip")

	// volume exists
	if utils.FileExists(volumeZipFile) {
		return volumeDir, nil
	}

	req := new(http.Request)
	req.SetUri(v.Meta.URL)
	res, err := http.Execute(req)
	if err != nil {
		return "", err
	}
	body := res.Body()
	defer body.Close()

	err = utils.WriteFile(volumeZipFile, body)
	if err != nil {
		os.RemoveAll(volumeDir)
		return "", err
	}

	volumeMD5, err := utils.CalculateFileMD5(volumeZipFile)
	if err != nil {
		os.RemoveAll(volumeDir)
		return "", err
	}
	if volumeMD5 != v.Meta.MD5 {
		os.RemoveAll(volumeDir)
		return "", fmt.Errorf("dateset (%s) downloaded with unexpected MD5", v.Name)
	}

	err = archiver.Zip.Open(volumeZipFile, volumeDir)
	if err != nil {
		os.RemoveAll(volumeDir)
		return "", err
	}
	return volumeDir, nil
}
