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
		// case config.ProofHostName:
		// todo get hostname
		// case config.ProofMachineID:
		// todo get MachineID
		// case config.ProofSystemUUID:
		// todo get SystemUUID
		default:
			return "", ErrProofTypeNotSupported
		}
	}
	return "", nil
}
