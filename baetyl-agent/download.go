package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

func (a *agent) downloadVolumes(volumes []baetyl.VolumeInfo) error {
	for _, v := range volumes {
		if v.Meta.URL == "" {
			continue
		}
		_, _, err := a.downloadVolume(v, "", true)
		if err != nil {
			return err
		}
	}
	return nil
}

func (a *agent) downloadVolume(v baetyl.VolumeInfo, name string, zip bool) (string, string, error) {
	rp, err := filepath.Rel(baetyl.DefaultDBDir, v.Path)
	if err != nil {
		return "", "", fmt.Errorf("path of volume (%s) invalid: %s", v.Name, err.Error())
	}
	hostDir := path.Join(baetyl.DefaultDBDir, rp)
	containerDir := path.Join(baetyl.DefaultDBDir, "volumes", rp)
	if name == "" {
		name = v.Name + ".zip"
	}
	containerFile := path.Join(containerDir, name)

	// volume exists
	if utils.FileExists(containerFile) {
		md5, err := utils.CalculateFileMD5(containerFile)
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
	err = utils.WriteFile(containerFile, res)
	if err != nil {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("failed to prepare volume (%s): %s", v.Name, err.Error())
	}

	md5, err := utils.CalculateFileMD5(containerFile)
	if err != nil {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("failed to calculate MD5 of volume (%s): %s", v.Name, err.Error())
	}
	if md5 != v.Meta.MD5 {
		os.RemoveAll(containerDir)
		return "", "", fmt.Errorf("MD5 of volume (%s) invalid", v.Name)
	}

	if zip {
		err = utils.Unzip(containerFile, containerDir)
		if err != nil {
			os.RemoveAll(containerDir)
			return "", "", fmt.Errorf("failed to unzip volume (%s): %s", v.Name, err.Error())
		}
	}
	return hostDir, containerDir, nil
}
