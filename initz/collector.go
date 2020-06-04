package initz

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/errors"
)

// TODO: can be configured by cloud
const (
	defaultSNPath = "var/lib/baetyl/sn"
)

// ErrProofTypeNotSupported the proof type is not supported
var ErrProofTypeNotSupported = fmt.Errorf("the proof type is not supported")

// ErrProofValueNotFound the proof value is not found
var ErrProofValueNotFound = fmt.Errorf("the proof value is not found")

func (init *Initialize) collect() (string, error) {
	fs := init.cfg.Init.ActivateConfig.Fingerprints
	if len(fs) == 0 {
		return "", nil
	}
	nodeInfo, err := init.ami.CollectNodeInfo()
	if err != nil {
		return "", errors.Trace(err)
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
				return "", errors.Trace(err)
			}
			return strings.TrimSpace(string(snByte)), nil
		case config.ProofHostName:
			if nodeInfo == nil {
				return "", errors.Trace(ErrProofValueNotFound)
			}
			return nodeInfo.Hostname, nil
		case config.ProofMachineID:
			if nodeInfo == nil {
				return "", errors.Trace(ErrProofValueNotFound)
			}
			return nodeInfo.MachineID, nil
		case config.ProofSystemUUID:
			if nodeInfo == nil {
				return "", errors.Trace(ErrProofValueNotFound)
			}
			return nodeInfo.SystemUUID, nil
		case config.ProofBootID:
			if nodeInfo == nil {
				return "", errors.Trace(ErrProofValueNotFound)
			}
			return nodeInfo.BootID, nil
		default:
			return "", errors.Trace(ErrProofTypeNotSupported)
		}
	}
	return "", nil
}
