package sync

import (
	"bytes"
	"time"

	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
	"github.com/pkg/errors"
)

func (s *sync) downloadObject(obj *specv1.ConfigurationObject, dir, name string, zip bool) error {
	// file exists
	if utils.FileExists(name) {
		md5, err := utils.CalculateFileMD5(name)
		if err == nil && md5 == obj.MD5 {
			s.log.Debug("file exists", log.Any("name", name))
			return nil
		}
	}

	headers := make(map[string]string)
	if obj.Metadata.Source == "baidubos" {
		headers["x-bce-security-token"] = obj.Metadata.Token
	}
	// TODO: streaming mode
	resp, err := s.http.GetURL(obj.URL, headers)
	if err != nil || resp == nil {
		// retry
		time.Sleep(time.Second)
		resp, err := s.http.GetURL(obj.URL, headers)
		if err != nil || resp == nil {
			return errors.Errorf("failed to download file (%s)", name)
		}
	}
	data, err := http.HandleResponse(resp)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return errors.WithStack(err)
	}

	err = utils.WriteFile(name, bytes.NewBuffer(data))
	if err != nil {
		return errors.Errorf("failed to prepare volume (%s): %s", name, err.Error())
	}

	if obj.MD5 != "" {
		md5, err := utils.CalculateFileMD5(name)
		if err != nil {
			return errors.Errorf("failed to calculate MD5 of volume (%s): %s", name, err.Error())
		}
		if md5 != obj.MD5 {
			return errors.Errorf("MD5 of volume (%s) invalid", name)
		}
	}

	if zip {
		err = utils.Unzip(name, dir)
		if err != nil {
			return errors.Errorf("failed to unzip file (%s): %s", name, err.Error())
		}
	}
	return nil
}
