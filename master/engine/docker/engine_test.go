package docker

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/engine"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	opt := engine.Options{
		Grace:      2 * time.Second,
		Pwd:        "/var/db",
		APIVersion: "1.38",
	}
	e, err := New(new(mockStats), opt)
	assert.NoError(t, err)
	assert.NotEmpty(t, e)
	defer e.Close()

	assert.Equal(t, NAME, e.Name())
	assert.Equal(t, "/var/db", e.(*dockerEngine).pwd)
	assert.Equal(t, 2*time.Second, e.(*dockerEngine).grace)
}

func TestRun(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Skip("docker not installed")
	}
	defer cli.Close()
	stats := mockStats{services: map[string]engine.InstancesStats{}}
	e := &dockerEngine{
		InfoStats: &stats,
		cli:       cli,
		networks:  map[string]string{},
		grace:     30 * time.Second,
		log:       newMockLogger(),
	}

	name := t.Name()
	err = e.removeContainerByName(name)
	assert.NoError(t, err)

	dir, _ := os.Getwd()
	sv := baetyl.ServiceVolume{
		Type:     "",
		Source:   path.Join(dir, "testrun/etc"),
		Target:   "/etc/baetyl",
		ReadOnly: true,
	}
	defer os.RemoveAll("testrun")

	vs := map[string]baetyl.ComposeVolume{}
	cs := baetyl.ComposeService{
		ContainerName: "",
		Hostname:      "",
		Image:         "hub.baidubce.com/baetyl-sandbox/baetyl-test:latest",
		Replica:       1,
		Volumes:       []baetyl.ServiceVolume{sv},
		NetworkMode:   "",
		Networks:      baetyl.NetworksInfo{},
		Ports:         []string{"12883:12883"},
		Devices:       nil,
		DependsOn:     nil,
		Command:       baetyl.Command{},
		Environment:   baetyl.Environment{},
		Restart:       baetyl.RestartPolicyInfo{},
		Resources:     baetyl.Resources{},
		Runtime:       "",
	}
	css := map[string]baetyl.ComposeService{}
	css["test"] = cs

	conf := baetyl.ComposeAppConfig{
		Version:    "3",
		AppVersion: "v0",
		Services:   css,
		Volumes:    map[string]baetyl.ComposeVolume{},
		Networks:   map[string]baetyl.ComposeNetwork{},
	}
	e.Prepare(conf)
	s, err := e.Run(name, cs, vs)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	ch := make(chan int, 1)
	go func() {
		for {
			s.Stats()
			stats.Lock()
			_, ok := stats.services[s.Name()]
			stats.Unlock()
			if ok {
				ch <- 1
				return
			}
		}
	}()
	<-ch
	s.Stop()
}

func TestRecover(t *testing.T) {
	cli, err := client.NewEnvClient()
	if err != nil {
		t.Skip("docker not installed")
	}
	defer cli.Close()
	stats := mockStats{services: map[string]engine.InstancesStats{}}
	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Container: struct {
					ID   string `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{},
			},
		},
	}
	logger := newMockLogger()
	e := &dockerEngine{
		InfoStats: &stats,
		cli:       cli,
		networks:  map[string]string{},
		grace:     30 * time.Second,
		log:       logger,
	}

	logger.records = []string{}
	e.Recover()
	assert.Equal(t, []string{"[Warnf][test][test] container id not found, maybe running mode changed"}, logger.records)

	id := "123412341234"
	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Container: struct {
					ID   string `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{
					ID: id,
				},
			},
		},
	}
	logger.records = []string{}
	e.Recover()
	expected := []string{
		fmt.Sprintf("[Debugf]container (%s) is stopping", id),
		fmt.Sprintf("[Warnf]failed to stop container (%s)", id),
		fmt.Sprintf("[Warnf][test][test] failed to stop the old container (%s)", id),
		fmt.Sprintf("[Warnf]failed to remove container (%s)", id),
		fmt.Sprintf("[Warnf][test][test] failed to remove the old container (%s)", id),
	}
	assert.Equal(t, expected, logger.records)

	name := t.Name()
	err = e.removeContainerByName(name)
	assert.NoError(t, err)
	image := "hub.baidubce.com/baetyl-sandbox/baetyl-test:latest"
	err = e.pullImage(image)
	assert.NoError(t, err)

	var params containerConfigs
	params.config = container.Config{
		Image: strings.TrimSpace(image),
	}
	id, err = e.startContainer(name, params)
	assert.NoError(t, err)
	assert.NotEmpty(t, id)

	ch := make(chan int, 1)
	go func() {
		for {
			stats := e.statsContainer(id)
			if stats["error"] == nil {
				ch <- 1
				return
			}
		}
	}()
	<-ch

	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Container: struct {
					ID   string `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{
					ID: id,
				},
			},
		},
	}
	logger.records = []string{}
	e.Recover()
	id = id[:12]
	expected = []string{
		fmt.Sprintf("[Debugf]container (%s) is stopping", id),
		fmt.Sprintf("[Debugf]container (%s) exit status: {<nil> 2}", id),
		fmt.Sprintf("[Infof][test][test] old container (%s) stopped", id),
		fmt.Sprintf("[Debugf]container (%s) removed", id),
		fmt.Sprintf("[Infof][test][test] old container (%s) removed", id),
	}
	assert.Equal(t, expected, logger.records)
}

func Test_parseDeviceSpecs(t *testing.T) {
	e := &dockerEngine{}
	var devices []string
	devices = append(devices, "/etc")
	devices = append(devices, "/etc:/etc/baetyl")
	devices = append(devices, "/etc:/etc/baetyl:mro")
	bindings, err := e.parseDeviceSpecs(devices)
	assert.NoError(t, err)
	assert.Equal(t, "/etc", bindings[0].PathOnHost)
	assert.Equal(t, "/etc", bindings[0].PathInContainer)
	assert.Equal(t, "mrw", bindings[0].CgroupPermissions)
	assert.Equal(t, "/etc", bindings[1].PathOnHost)
	assert.Equal(t, "/etc/baetyl", bindings[1].PathInContainer)
	assert.Equal(t, "mrw", bindings[1].CgroupPermissions)
	assert.Equal(t, "/etc", bindings[2].PathOnHost)
	assert.Equal(t, "/etc/baetyl", bindings[2].PathInContainer)
	assert.Equal(t, "mro", bindings[2].CgroupPermissions)
}

type mockStats struct {
	stats    map[string]map[string]attribute
	services map[string]engine.InstancesStats
	sync.Mutex
}

func (s *mockStats) LoadStats(sss interface{}) bool {
	s.Lock()
	defer s.Unlock()
	var buf bytes.Buffer
	gob.NewEncoder(&buf).Encode(s.stats)
	gob.NewDecoder(bytes.NewBuffer(buf.Bytes())).Decode(sss)
	return true
}

func (s *mockStats) SetInstanceStats(serviceName, instanceName string, partialStats engine.PartialStats, persist bool) {
	s.Lock()
	defer s.Unlock()
	service, ok := s.services[serviceName]
	if !ok {
		service = engine.InstancesStats{}
		s.services[serviceName] = service
	}
	instance, ok := service[instanceName]
	if !ok {
		instance = partialStats
		service[instanceName] = instance
	} else {
		for k, v := range partialStats {
			instance[k] = v
		}
	}
}

func (*mockStats) DelInstanceStats(serviceName, instanceName string, persist bool) {
	return
}

func (*mockStats) DelServiceStats(serviceName string, persist bool) {
	return
}

type mockLogger struct {
	records []string
	data    map[string]interface{}
	err     error
	sync.Mutex
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		data:    map[string]interface{}{},
		records: []string{},
	}
}

func (l *mockLogger) WithField(key string, value interface{}) logger.Logger {
	l.Lock()
	defer l.Unlock()
	l.data[key] = value
	return l
}
func (l *mockLogger) WithError(err error) logger.Logger {
	l.Lock()
	defer l.Unlock()
	l.err = err
	return l
}
func (l *mockLogger) Debugf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Debugf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Infof(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Infof]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Warnf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Warnf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Errorf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Errorf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Fatalf(format string, args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Fatalf]"+fmt.Sprintf(format, args...))
}
func (l *mockLogger) Debugln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Debugln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Infoln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Infoln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Warnln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Warnln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Errorln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Errorln]"+fmt.Sprintln(args...))
}
func (l *mockLogger) Fatalln(args ...interface{}) {
	l.Lock()
	defer l.Unlock()
	l.records = append(l.records, "[Fatalln]"+fmt.Sprintln(args...))
}
