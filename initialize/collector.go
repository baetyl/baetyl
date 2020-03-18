package initialize

import (
	"github.com/baetyl/baetyl-core/common"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

func (init *initialize) collect() (string, error) {
	fs := init.cfg.Init.ActivateConfig.Fingerprints
	if fs == nil || len(fs) == 0 {
		return "", nil
	}
	for _, f := range fs {
		switch f.Proof {
		case common.Input:
			if init.attrs != nil {
				return init.attrs[f.Value], nil
			}
		case common.SN:
			snByte, err := ioutil.ReadFile(path.Join(common.DefaultSNPath, f.Value))
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(snByte)), nil
		case common.HostName:
			return os.Getenv(common.KeyHostName), nil
		case common.MachineID:
			// todo get MachineID
		case common.SystemUUID:
			// todo get SystemUUID
		}
	}
	return "", nil
}
