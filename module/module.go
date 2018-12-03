package module

import (
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/baidu/openedge/utils"
	"github.com/juju/errors"
)

// Module module interfaces
type Module interface {
	Start() error
	Close()
}

// Load load a module config
func Load(confObject interface{}, conf string) error {
	var err error
	var confBytes []byte
	conf = strings.TrimSpace(conf)
	unmarshal := utils.UnmarshalYAML
	if strings.HasPrefix(conf, "{") && strings.HasSuffix(conf, "}") {
		confBytes = []byte(conf)
		unmarshal = utils.UnmarshalJSON
	} else {
		confBytes, err = ioutil.ReadFile(conf)
		if err != nil {
			return errors.Trace(err)
		}
	}
	return errors.Trace(unmarshal(confBytes, confObject))
}

// Wait waits until module exit
func Wait() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	signal.Ignore(syscall.SIGPIPE)
	<-sig
}
