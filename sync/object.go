package sync

import (
	"encoding/json"
	"io"
	gohttp "net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
)

func FilterConfig(cfg *specv1.Configuration) {
	proLabel, ok := cfg.Labels["baetyl-config-type"]
	isProgram := ok && proLabel == "baetyl-program"
	if !isProgram {
		return
	}
	mode := context.RunMode()
	platform := context.PlatformString()
	for k, _ := range cfg.Data {
		if !specv1.IsConfigObject(k) {
			continue
		}
		if mode == context.RunModeNative && strings.Contains(k, platform) {
			continue
		}
		delete(cfg.Data, k)
	}
}

func DownloadConfig(cli *http.Client, objectPath string, cfg *specv1.Configuration) error {
	for k, v := range cfg.Data {
		if !specv1.IsConfigObject(k) {
			continue
		}

		dir := filepath.Join(objectPath, cfg.Name, cfg.Version)
		err := os.MkdirAll(dir, 0755)
		if err != nil {
			return errors.Errorf("failed to prepare path of config object (%s): %s", cfg.Name, err)
		}

		obj := new(specv1.ConfigurationObject)
		err = json.Unmarshal([]byte(v), &obj)
		if err != nil {
			log.L().Warn("failed to unmarshal config object", log.Any("name", cfg.Name), log.Any("key", k), log.Error(err))
			return errors.Errorf("failed to unmarshal config object (%s): %s", cfg.Name, err)
		}
		filename := filepath.Join(dir, strings.TrimPrefix(k, specv1.PrefixConfigObject))
		err = downloadObject(cli, obj, dir, filename, obj.Unpack)
		if err != nil {
			os.RemoveAll(dir)
			return errors.Trace(err)
		}
		if hook, ok := Hooks[BaetylHookUploadObject]; ok {
			if roam, ok := hook.(UploadObjectFunc); ok {
				log.L().Info("upload file to worker node", log.Any("file", filename))
				err := roam(dir, filename, obj.MD5, obj.Unpack)
				if err != nil {
					log.L().Warn("failed to upload file to node", log.Any("file", filename))
					return errors.Trace(err)
				}
			}
		}
	}
	return nil
}

func downloadObject(cli *http.Client, obj *specv1.ConfigurationObject, dir, name, unpack string) error {
	lockfile, err := os.OpenFile(name+".baetyl-lock", os.O_CREATE|os.O_RDWR, 0755)
	if err != nil {
		return err
	}
	if err = utils.Flock(lockfile, 0); err != nil {
		return err
	}
	clean := func() {
		utils.Funlock(lockfile)
		os.Remove(lockfile.Name())
	}
	defer clean()
	if obj.MD5 != "" {
		md5, err := utils.CalculateFileMD5(name)
		if err == nil && md5 == obj.MD5 {
			log.L().Debug("config object file exists", log.Any("name", name))
			return nil
		}
	} else {
		if utils.FileExists(name) {
			log.L().Debug("config object file exists", log.Any("name", name))
			return nil
		}
	}

	headers := make(map[string]string)
	if obj.Token != "" {
		headers["x-bce-security-token"] = obj.Token
	}
	resp, err := cli.GetURL(obj.URL, headers)
	if err != nil || resp == nil {
		// retry
		time.Sleep(time.Second)
		resp, err := cli.GetURL(obj.URL, headers)
		if err != nil || resp == nil {
			return errors.Errorf("failed to download config object (%s): %v", name, err)
		}
	}
	if resp.StatusCode != gohttp.StatusOK {
		return errors.Errorf("failed to download config object (%s): [%d] %s", name, resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0755)
	if err := file.Truncate(0); err != nil {
		return errors.Trace(err)
	}
	if _, err = io.Copy(file, resp.Body); err != nil {
		log.L().Error("failed to download config object file", log.Error(err))
		return errors.Errorf("failed to download config object file (%s): %v", name, err)
	}

	if obj.MD5 != "" {
		md5, err := utils.CalculateFileMD5(name)
		if err != nil {
			return errors.Errorf("failed to calculate MD5 of config object (%s): %s", name, err.Error())
		}
		if md5 != obj.MD5 {
			return errors.Errorf("MD5 of config object (%s) invalid", name)
		}
	}

	switch unpack {
	case "":
	case "zip":
		err = utils.Unzip(name, dir)
		if err != nil {
			return errors.Errorf("failed to unzip file (%s): %s", name, err.Error())
		}
	default:
		return errors.Errorf("failed to unpack file (%s): '%s' not supported", name, unpack)
	}
	return nil
}
