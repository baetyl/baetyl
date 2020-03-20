package docker

func (s *service) runPod(name string, template *template, stop chan struct{}) {
	cid, err := s.e.startContainer(name, template)
	if err != nil {
		s.log.WithError(err).Warnln("failed to start instance, clean and retry")
		// remove and retry
		s.e.removeContainerByName(name)
		cid, err = s.e.startContainer(name, template)
		if err != nil {
			s.e.removeContainerByName(name)
			s.log.WithError(err).Warnln("failed to start instance again")
			<-stop
			s.wg.Done()
			return
		}
	}
	exit := make(chan error, 1)
	restart := true
	for restart {
		go func() {
			exit <- s.e.waitContainer(cid)
		}()
		select {
		case err = <-exit:
			if err != nil {
				s.log.Infof("container stopped name=%s error=%s", name, err.Error())
			}
			err = s.e.restartContainer(cid)
			if err != nil {
				s.log.Warnln("restart container fail (%s)", name)
				s.e.removeContainer(cid)
				<-stop
				goto EXIT
			}
		case <-stop:
			s.e.stopContainer(cid)
			<-exit
			s.e.removeContainer(cid)
			goto EXIT
		}
	}
EXIT:
	s.wg.Done()
	return
}
