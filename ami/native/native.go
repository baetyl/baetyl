package native

import (
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl/ami"
	"github.com/baetyl/baetyl/config"
)

type nativeImpl struct {
	log *log.Logger
}

func init() {
	ami.Register("native", newNativeImpl)
}

func newNativeImpl(cfg config.AmiConfig) (ami.AMI, error) {
	return nil, nil
}
