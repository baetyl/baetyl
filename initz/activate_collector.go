package initz

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/baetyl/baetyl-go/v2/errors"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"

	"github.com/baetyl/baetyl/v2/config"
)

// TODO: can be configured by cloud
const (
	defaultSNPath = "var/lib/baetyl/sn"
)

var (
	// ErrProofTypeNotSupported the proof type is not supported
	ErrProofTypeNotSupported = fmt.Errorf("the proof type is not supported")
	// ErrProofValueNotFound the proof value is not found
	ErrProofValueNotFound = fmt.Errorf("the proof value is not found")
	// ErrProofValueNotFound the proof value is not found
	ErrGetMasterNodeInfo = fmt.Errorf("failed to get master node info")
)

func (active *Activate) collect() (string, error) {
	fs := active.cfg.Init.Active.Collector.Fingerprints
	if len(fs) == 0 {
		return "", nil
	}
	infos, err := active.ami.CollectNodeInfo()
	if err != nil {
		return "", errors.Trace(err)
	}
	info, ok := infos[active.ami.GetMasterNodeName()]
	if !ok {
		return "", errors.Trace(ErrGetMasterNodeInfo)
	}
	nodeInfo, ok := info.(*specV1.NodeInfo)
	if !ok {
		return "", errors.Trace(ErrGetMasterNodeInfo)
	}
	for _, f := range fs {
		switch f.Proof {
		case config.ProofInput:
			if active.attrs != nil {
				return active.attrs[f.Value], nil
			}
		case config.ProofSN:
			snByte, err := ioutil.ReadFile(path.Join(defaultSNPath, f.Value))
			if err != nil {
				return "", errors.Trace(err)
			}
			return strings.TrimSpace(string(snByte)), nil
		case config.ProofHostName:
			return nodeInfo.Hostname, nil
		case config.ProofMachineID:
			return nodeInfo.MachineID, nil
		case config.ProofSystemUUID:
			return nodeInfo.SystemUUID, nil
		case config.ProofBootID:
			return nodeInfo.BootID, nil
		default:
			return "", errors.Trace(ErrProofTypeNotSupported)
		}
	}
	return "", nil
}
