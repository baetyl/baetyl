package docker

import (
	"fmt"
	"sync"

	"github.com/baetyl/baetyl/logger"
	schema "github.com/baetyl/baetyl/schema/v3"
	cmap "github.com/orcaman/concurrent-map"
)

const (
	fmtVolumeRW = "%s:%s:rw"
	fmtVolumeRO = "%s:%s:ro"
)

type service struct {
	e        *engine
	log      logger.Logger
	name     string
	meta     *schema.ComposeService
	template template
	pods     cmap.ConcurrentMap
	wg       sync.WaitGroup
	sig      chan struct{}
}

func (s *service) run() {
	s.log.Debugf("%s replica: %s, %d", s.name, s.meta.Image, s.meta.Replica)
	var pid uint64 = 0
	for i := 0; i < s.meta.Replica; i++ {
		select {
		case <-s.sig:
			goto CLEAN
		default:
			break
		}
		s.wg.Add(1)
		stop := make(chan struct{}, 1)
		pname := fmt.Sprintf("%s.%d", s.name, pid)
		s.pods.Set(pname, stop)
		go s.runPod(pname, &s.template, stop)
		pid = pid + 1
	}
	<-s.sig
CLEAN:
	for item := range s.pods.Iter() {
		item.Val.(chan struct{}) <- struct{}{}
	}
}

func (s *service) stop() {
	s.sig <- struct{}{}
	s.wg.Wait()
	s.log.Infof("service (%s) stopped", s.name)
}
