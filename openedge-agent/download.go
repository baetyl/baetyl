package main

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
	"github.com/mholt/archiver"
)

func (m *mo) downloadConfigVolume(cfgVol openedge.VolumeInfo) (*openedge.AppConfig, string, error) {
	volumeHostDir, volumeContainerDir, err := m.download(cfgVol)
	if err != nil {
		return nil, "", err
	}
	var cfg openedge.AppConfig
	cfgFile := path.Join(volumeContainerDir, openedge.AppConfFileName)
	err = utils.LoadYAML(cfgFile, &cfg)
	return &cfg, volumeHostDir, err
}

func (m *mo) downloadAppVolumes(cfg *openedge.AppConfig) error {
	for _, ds := range cfg.Volumes {
		if ds.Meta.URL == "" {
			continue
		}
		_, _, err := m.download(ds)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *mo) download(v openedge.VolumeInfo) (string, string, error) {
	rp, err := filepath.Rel(openedge.DefaultDBDir, v.Path)
	if err != nil {
		return "", "", fmt.Errorf("path of volume (%s) invalid: %s", v.Name, err.Error())
	}

	volumeHostDir := path.Join(openedge.DefaultDBDir, rp)
	volumeContainerDir := path.Join(openedge.DefaultDBDir, "volumes", rp)
	volumeZipFile := path.Join(volumeContainerDir, v.Name+".zip")

	// volume exists
	if utils.FileExists(volumeZipFile) {
		volumeMD5, err := utils.CalculateFileMD5(volumeZipFile)
		if err == nil && volumeMD5 == v.Meta.MD5 {
			m.ctx.Log().Debugf("volume (%s) exists", v.Name)
			return volumeHostDir, volumeContainerDir, nil
		}
	}

	res, err := m.http.SendUrl("GET", v.Meta.URL, nil, nil)
	if err != nil || res == nil {
		// retry
		time.Sleep(time.Second)
		res, err = m.http.SendUrl("GET", v.Meta.URL, nil, nil)
		if err != nil || res == nil {
			return "", "", fmt.Errorf("failed to download volume (%s): %v", v.Name, err)
		}
	}
	defer res.Close()

	err = os.MkdirAll(volumeContainerDir, 0755)
	if err != nil {
		return "", "", fmt.Errorf("failed to prepare volume (%s): %s", v.Name, err.Error())
	}
	err = utils.WriteFile(volumeZipFile, res)
	if err != nil {
		os.RemoveAll(volumeContainerDir)
		return "", "", fmt.Errorf("failed to prepare volume (%s): %s", v.Name, err.Error())
	}

	volumeMD5, err := utils.CalculateFileMD5(volumeZipFile)
	if err != nil {
		os.RemoveAll(volumeContainerDir)
		return "", "", fmt.Errorf("failed to calculate MD5 of volume (%s): %s", v.Name, err.Error())
	}
	if volumeMD5 != v.Meta.MD5 {
		os.RemoveAll(volumeContainerDir)
		return "", "", fmt.Errorf("MD5 of volume (%s) invalid", v.Name)
	}

	err = archiver.Zip.Open(volumeZipFile, volumeContainerDir)
	if err != nil {
		os.RemoveAll(volumeContainerDir)
		return "", "", fmt.Errorf("failed to unzip volume (%s): %s", v.Name, err.Error())
	}
	return volumeHostDir, volumeContainerDir, nil
}
