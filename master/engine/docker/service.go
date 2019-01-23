package engine

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"time"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	"github.com/docker/docker/api/types"
)

type dockerService struct {
	e      *dockerEngine
	id     string
	si     *openedge.ServiceInfo
	cfgdir string
	rmcfg  bool
}

func (s *dockerService) Info() *openedge.ServiceInfo {
	return s.si
}

func (s *dockerService) Instances() []engine.Instance {
	return []engine.Instance{}
}

func (s *dockerService) Scale(replica int, grace time.Duration) error {
	return errors.New("not implemented yet")
}

func (s *dockerService) Stats() openedge.ServiceStatus {
	instance := openedge.InstanceStatus{"id": s.id}
	result := openedge.NewServiceStatus(s.si.Name)
	result.Instances = append(result.Instances, instance)

	ctx := context.Background()
	iresp, err := s.e.client.ContainerInspect(ctx, s.id)
	if err != nil {
		instance["error"] = err.Error()
		return result
	}
	sresp, err := s.e.client.ContainerStats(ctx, s.id, false)
	if err != nil {
		instance["error"] = err.Error()
		return result
	}
	defer sresp.Body.Close()
	data, err := ioutil.ReadAll(sresp.Body)
	if err != nil {
		instance["error"] = err.Error()
		return result
	}
	var tstats types.Stats
	err = json.Unmarshal(data, &tstats)
	if err != nil {
		instance["error"] = err.Error()
		return result
	}

	instance["status"] = iresp.State.Status
	instance["start_time"] = iresp.State.StartedAt
	instance["finish_time"] = iresp.State.FinishedAt
	instance["cpu_stats"] = tstats.CPUStats
	instance["memory_stats"] = tstats.MemoryStats
	return result
}

func (s *dockerService) Stop(grace time.Duration) error {
	if err := s.e.client.ContainerStop(context.Background(), s.id, &grace); err != nil {
		openedge.Errorln("failed to stop container:", err.Error())
	}
	err := s.e.client.ContainerRemove(context.Background(), s.id, types.ContainerRemoveOptions{Force: true})
	if s.rmcfg {
		os.RemoveAll(s.cfgdir)
	}
	return err
}
