package sync

import (
	"io"
	gohttp "net/http"
	"os"
	"syscall"
	"time"

	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/log"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
	"github.com/baetyl/baetyl-go/utils"
)

const (
	flockRetryTimeout = time.Microsecond * 100
)

func (s *sync) downloadObject(obj *specv1.ConfigurationObject, dir, name string, zip bool) error {
	file, err := os.OpenFile(name, os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return err
	}
	if err = flock(file, 0); err != nil {
		return err
	}
	defer funlock(file)
	if obj.MD5 == "" {
		return nil
	}
	md5, err := utils.CalculateFileMD5(name)
	if err == nil && md5 == obj.MD5 {
		s.log.Debug("file exists", log.Any("name", name))
		return nil
	}

	headers := make(map[string]string)
	if obj.Token != "" {
		headers["x-bce-security-token"] = obj.Token
	}
	resp, err := s.http.GetURL(obj.URL, headers)
	if err != nil || resp == nil {
		// retry
		time.Sleep(time.Second)
		resp, err := s.http.GetURL(obj.URL, headers)
		if err != nil || resp == nil {
			return errors.Errorf("failed to download file (%s)", name)
		}
	}
	if resp.StatusCode != gohttp.StatusOK {
		return errors.Errorf("[%d] %s", resp.StatusCode, resp.Status)
	}
	defer resp.Body.Close()
	if err := file.Truncate(0); err != nil {
		return errors.Trace(err)
	}
	if _, err = io.Copy(file, resp.Body); err != nil {
		s.log.Error("failed to download data", log.Error(err))
		return errors.Trace(err)
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

// only works on unix
func flock(file *os.File, timeout time.Duration) error {
	var t time.Time
	if timeout != 0 {
		t = time.Now()
	}
	fd := file.Fd()
	flag := syscall.LOCK_NB | syscall.LOCK_EX
	for {
		err := syscall.Flock(int(fd), flag)
		if err == nil {
			return nil
		} else if err != syscall.EWOULDBLOCK {
			return err
		}
		if timeout != 0 && time.Since(t) > timeout-flockRetryTimeout {
			return errors.Errorf("time out")
		}
		time.Sleep(flockRetryTimeout)
	}
}

func funlock(file *os.File) error {
	return syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
}
