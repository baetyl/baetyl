package initialize

import (
	"errors"
	"io/ioutil"
	"path"
	"strings"

	"github.com/baetyl/baetyl-core/config"
	v1 "github.com/baetyl/baetyl-go/spec/v1"
)

// TODO: can be configured by cloud
const (
	defaultSNPath = "var/lib/baetyl/sn"
)

// ErrProofTypeNotSupported the proof type is not supported
var ErrProofTypeNotSupported = errors.New("the proof type is not supported")

// ErrProofValueNotFound the proof value is not found
var ErrProofValueNotFound = errors.New("the proof value is not found")

func (init *Initialize) collect() (string, error) {
	fs := init.cfg.Init.ActivateConfig.Fingerprints
	if len(fs) == 0 {
		return "", nil
	}
	report, err := init.ami.CollectInfo()
	if err != nil {
		return "", err
	}
	nodeInfo := report["node"]
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
			if nodeInfo == nil {
				return "", ErrProofValueNotFound
			}
			return nodeInfo.(v1.NodeInfo).Hostname, nil
		case config.ProofMachineID:
			if nodeInfo == nil {
				return "", ErrProofValueNotFound
			}
			return nodeInfo.(v1.NodeInfo).MachineID, nil
		case config.ProofSystemUUID:
			if nodeInfo == nil {
				return "", ErrProofValueNotFound
			}
			return nodeInfo.(v1.NodeInfo).SystemUUID, nil
		case config.ProofBootID:
			if nodeInfo == nil {
				return "", ErrProofValueNotFound
			}
			return nodeInfo.(v1.NodeInfo).BootID, nil
		default:
			return "", ErrProofTypeNotSupported
		}
	}
	return "", nil
}
