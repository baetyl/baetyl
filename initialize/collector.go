package initialize

import (
	"errors"
	"io/ioutil"
	"path"
	"strings"

	"github.com/baetyl/baetyl-core/config"
)

// TODO: can be configured by cloud
const (
	defaultSNPath = "var/lib/baetyl/sn"
)

// ErrProofTypeNotSupported the proof type is not supported
var ErrProofTypeNotSupported = errors.New("the proof type is not supported")

func (init *Initialize) collect() (string, error) {
	fs := init.cfg.Init.ActivateConfig.Fingerprints
	if len(fs) == 0 {
		return "", nil
	}
	report, err := init.ami.CollectInfo()
	if err != nil {
		return "", err
	}
	for _, f := range fs {
		switch f.Proof {
		case config.ProofInput:
			if init.attrs != nil {
				return init.attrs[f.Value], nil
			}
		case config.ProofSN:
			snByte, err := ioutil.ReadFile(path.Join(defaultSNPath, f.Value))
			if err != nil {
				return "", err
			}
			return strings.TrimSpace(string(snByte)), nil
		case config.ProofHostName:
			return report.NodeInfo.Hostname, nil
		case config.ProofMachineID:
			return report.NodeInfo.MachineID, nil
		case config.ProofSystemUUID:
			return report.NodeInfo.SystemUUID, nil
		case config.ProofBootID:
			return report.NodeInfo.BootID, nil
		default:
			return "", ErrProofTypeNotSupported
		}
	}
	return "", nil
}
