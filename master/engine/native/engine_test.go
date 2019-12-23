package native

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/master/engine"
	"github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/shirou/gopsutil/process"
	"github.com/stretchr/testify/assert"
)

func Test_mount(t *testing.T) {
	epwd, err := ioutil.TempDir("", t.Name())
	assert.NoError(t, err)
	defer os.RemoveAll(epwd)
	srcdir := path.Join(epwd, "srcdir")
	srcfile := path.Join(epwd, "srcfile")
	srcnone := path.Join(epwd, "srcnone")
	srcerr := path.Join(epwd, "srcfile/srcnone")
	err = os.MkdirAll(srcdir, 0700)
	assert.NoError(t, err)
	err = ioutil.WriteFile(srcfile, []byte(""), 0700)
	assert.NoError(t, err)
	dstdir := path.Join(epwd, "dstdir")
	dstfile := path.Join(epwd, "dstfile")
	dstnone := path.Join(epwd, "dstnone")
	dsterr := path.Join(epwd, "dstfile/dstnone")

	type args struct {
		src string
		dst string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "file",
			args: args{
				src: srcfile,
				dst: dstfile,
			},
		},
		{
			name: "dir1",
			args: args{
				src: srcdir,
				dst: dstdir,
			},
		},
		{
			name: "dir2",
			args: args{
				src: srcnone,
				dst: dstnone,
			},
		},
		{
			name: "err1",
			args: args{
				src: srcerr,
				dst: dsterr,
			},
			wantErr: true,
		},
		{
			name: "err2",
			args: args{
				src: srcfile,
				dst: srcfile,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mount(tt.args.src, tt.args.dst)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("mount() error = %v, wantErr %v", err, tt.wantErr)
				}
			} else {
				wfile := tt.args.dst
				rfile := tt.args.src
				if strings.HasPrefix(tt.name, "dir") {
					wfile = path.Join(wfile, "tmpfile")
					rfile = path.Join(rfile, "tmpfile")
				}
				err = ioutil.WriteFile(wfile, []byte(epwd), 0700)
				assert.NoError(t, err)
				out, err := ioutil.ReadFile(rfile)
				assert.NoError(t, err)
				assert.Equal(t, []byte(epwd), out)
			}
		})
	}
}

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
	assert.Equal(t, "/var/db", e.(*nativeEngine).pwd)
	assert.Equal(t, 2*time.Second, e.(*nativeEngine).grace)
}

func TestNative(t *testing.T) {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.RemoveAll(path.Join(pwd, "testdata", "var", "run"))

	name := t.Name()
	stats := mockStats{services: map[string]engine.InstancesStats{}}
	e := &nativeEngine{
		pwd:       path.Join(pwd, "testdata"),
		grace:     10 * time.Second,
		InfoStats: &stats,
		log:       newMockLogger(),
	}
	defer e.Close()
	sv := baetyl.ServiceVolume{
		Source: "var/db/baetyl/cmd",
		Target: "/lib/baetyl/cmd",
	}
	cmd := baetyl.ComposeService{
		Image:       "cmd",
		Replica:     1,
		Volumes:     []baetyl.ServiceVolume{sv},
		Networks:    baetyl.NetworksInfo{},
		Ports:       []string{"13883:13883"},
		Command:     baetyl.Command{},
		Environment: baetyl.Environment{},
		Restart:     baetyl.RestartPolicyInfo{},
		Resources:   baetyl.Resources{},
	}
	s, err := e.Run(name, cmd, nil)
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

	sv = baetyl.ServiceVolume{
		Source: "var/db/baetyl/unknown/cmd",
		Target: "/lib/baetyl/cmd",
	}
	cmd = baetyl.ComposeService{
		Image:       "cmd",
		Replica:     1,
		Volumes:     []baetyl.ServiceVolume{sv},
		Networks:    baetyl.NetworksInfo{},
		Ports:       []string{"14883:14883"},
		Command:     baetyl.Command{},
		Environment: baetyl.Environment{},
		Restart:     baetyl.RestartPolicyInfo{},
		Resources:   baetyl.Resources{},
	}
	s, err = e.Run(name, cmd, nil)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
	fmt.Println(err)
}

func TestRecover(t *testing.T) {
	pwd, err := os.Getwd()
	assert.NoError(t, err)
	defer os.RemoveAll(path.Join(pwd, "testdata", "var", "run"))

	stats := mockStats{services: map[string]engine.InstancesStats{}}
	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Process: struct {
					ID   int    `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{},
			},
		},
	}
	logger := newMockLogger()
	e := &nativeEngine{
		pwd:       path.Join(pwd, "testdata"),
		grace:     10 * time.Second,
		InfoStats: &stats,
		log:       logger,
	}
	defer e.Close()
	logger.records = []string{}
	e.Recover()
	assert.Equal(t, []string{"[Warnf][test][test] process id not found, maybe running mode changed"}, logger.records)

	var pid int
	for {
		rand.Seed(time.Now().Unix())
		a := rand.Intn(100) + 50000
		_, err := process.NewProcess(int32(a))
		if err != nil {
			pid = a
			break
		}
	}
	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Process: struct {
					ID   int    `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{
					ID: pid,
				},
			},
		},
	}
	logger.records = []string{}
	e.Recover()
	assert.Equal(t, []string{fmt.Sprintf("[Warnf][test][test] failed to get old process (%d)", pid)}, logger.records)

	sh, err := exec.LookPath("sh")
	assert.NoError(t, err)
	var procAttr os.ProcAttr
	procAttr.Files = []*os.File{os.Stdin, os.Stdout, os.Stderr}
	pid2, err := os.StartProcess(sh, nil, &procAttr)
	assert.NoError(t, err)

	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Process: struct {
					ID   int    `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{
					ID:   pid2.Pid,
					Name: "test",
				},
			},
		},
	}
	logger.records = []string{}
	e.Recover()
	assert.Equal(t, []string{fmt.Sprintf("[Debugf][test][test] name of old process (%d) not matched, test -> sh", pid2.Pid)}, logger.records)

	stats.stats = map[string]map[string]attribute{
		"test": {
			"test": {
				Name: "",
				Process: struct {
					ID   int    `yaml:"id" json:"id"`
					Name string `yaml:"name" json:"name"`
				}{
					ID:   pid2.Pid,
					Name: "sh",
				},
			},
		},
	}
	logger.records = []string{}
	e.Recover()
	assert.Equal(t, []string{fmt.Sprintf("[Infof][test][test] old process (%d) stopped", pid2.Pid)}, logger.records)
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
