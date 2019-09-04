package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
	"github.com/mholt/archiver"
)

func (a *agent) downloadVolumes(volumes []baetyl.VolumeInfo) error {
	for _, v := range volumes {
		if v.Meta.URL == "" {
			continue
		}
		_, _, err := a.downloadVolume(v)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *agent) downloadVolume(v baetyl.VolumeInfo) (string, string, error) {
	rp, err := filepath.Rel(baetyl.DefaultDBDir, v.Path)
	if err != nil {
		return "", "", fmt.Errorf("path of volume (%s) invalid: %s", v.Name, err.Error())
	}

	hostDir := path.Join(baetyl.DefaultDBDir, rp)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	containerZipFile := path.Join(containerDir, v.Name+".zip")

	// volume exists
	if utils.FileExists(containerZipFile) {
		md5, err := utils.CalculateFileMD5(containerZipFile)
		if err == nil && md5 == v.Meta.MD5 {
			a.ctx.Log().Debugf("volume (%s) exists", v.Name)
			return hostDir, containerDir, nil
		}
	}

	res, err := a.http.SendUrl("GET", v.Meta.URL, nil, nil)
	if err != nil || res == nil {
		// retry
		time.Sleep(time.Second)
		res, err = a.http.SendUrl("GET", v.Meta.URL, nil, nil)
		if err != nil || res == nil {
			return "", "", fmt.Errorf("failed to download volume (%s): %v", v.Name, err)
		}
	}
	defer res.Close()

	err = os.MkdirAll(containerDir, 0755)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare volume (%s): %s", v.Name, err.Error())
	}
	err = utils.WriteFile(containerZipFile, res)
	if err != nil {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("failed to prepare volume (%s): %s", v.Name, err.Error())
	}

	md5, err := utils.CalculateFileMD5(containerZipFile)
	if err != nil {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("failed to calculate MD5 of volume (%s): %s", v.Name, err.Error())
	}
	if md5 != v.Meta.MD5 {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("MD5 of volume (%s) invalid", v.Name)
	}

	err = archiver.Zip.Open(containerZipFile, containerDir)
	if err != nil {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("failed to unzip volume (%s): %s", v.Name, err.Error())
	}
	return hostDir, containerDir, nil
}
