package sync

import (
	"bytes"
	"fmt"
	"time"

	"github.com/baetyl/baetyl-go/http"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/spec/api"
	"github.com/baetyl/baetyl-go/utils"
)

func (s *Sync) downloadObject(obj *api.CRDConfigObject, dir, name string, zip bool) error {
	// file exists
	if utils.FileExists(name) {
		md5, err := utils.CalculateFileMD5(name)
		if err == nil && md5 == obj.MD5 {
			s.log.Debug("file exists", log.Any("name", name))
			return nil
		}
	}

	// TODO: streaming mode
	resp, err := s.http.Get(obj.URL)
	if err != nil || resp == nil {
		// retry
		time.Sleep(time.Second)
		resp, err = s.http.Get(obj.URL)
		if err != nil || resp == nil {
			return fmt.Errorf("failed to download file (%s)", name)
		}
	}
	data, err := http.HandleResponse(resp)
	if err != nil {
		s.log.Error("failed to send report data", log.Error(err))
		return err
	}

	err = utils.WriteFile(name, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to prepare volume (%s): %s", name, err.Error())
	}

	md5, err := utils.CalculateFileMD5(name)
	if err != nil {
		return fmt.Errorf("failed to calculate MD5 of volume (%s): %s", name, err.Error())
	}
	if md5 != obj.MD5 {
		return fmt.Errorf("MD5 of volume (%s) invalid", name)
	}

	if zip {
		err = utils.Unzip(name, dir)
		if err != nil {
			return fmt.Errorf("failed to unzip file (%s): %s", name, err.Error())
		}
	}
	return nil
}
