package initialize

import (
	"github.com/baetyl/baetyl-core/common"
	"io/ioutil"
	"path"
	"strings"
)

const (
	DefaultSNPath = "var/lib/baetyl/sn"
)

func (init *Initialize) collect() (string, error) {
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
			snByte, err := ioutil.ReadFile(path.Join(DefaultSNPath, f.Value))
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(snByte)), nil
		case common.HostName:
			// todo get hostname
		case common.MachineID:
			// todo get MachineID
		case common.SystemUUID:
			// todo get SystemUUID
		}
	}
	return "", nil
}
