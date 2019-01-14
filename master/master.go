package master

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	cmap "github.com/orcaman/concurrent-map"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	sdk "github.com/baidu/openedge/sdk/go"
	"github.com/baidu/openedge/utils"
)

// Version of openedge
const Version = "version"

// Master master manages all modules and connects with cloud
type Master struct {
	cfg    Config
	engine engine.Engine
	dyncfg *DynamicConfig
	svcs   cmap.ConcurrentMap
	server Server
}

// New creates a new master
func New(workdir string, confpath string) (*Master, error) {
	data, err := ioutil.ReadFile(path.Join(workdir, confpath))
	if err != nil {
		return nil, err
	}
	m := &Master{
		svcs: cmap.New(),
	}
	err = utils.UnmarshalYAML(data, &m.cfg)
	if err != nil {
		return nil, err
	}
	err = defaults(&m.cfg)
	if err != nil {
		return nil, err
	}
	err = sdk.InitLogger(&m.cfg.Logger, "openedge", "master")
	if err != nil {
		return nil, err
	}
	openedge.Debugln("work dir:", workdir)
	m.engine, err = engine.New(m.cfg.Mode, workdir)
	if err != nil {
		return nil, err
	}
	err = m.server.start(m)
	if err != nil {
		m.engine.Close()
		return nil, err
	}
	err = m.Reload(path.Join(workdir, "var", "db", "openedge", "service"))
	if err != nil {
		m.Close()
		return nil, err
	}
	return m, nil
}

// Close closes agent
func (m *Master) Close() error {
	m.server.stop()
	m.cleanServices()
	m.engine.Close()
	return nil
}

/*
func (m *Master) authModule(username, password string) bool {
	return m.engine.Authenticate(username, password)
}

func (m *Master) startModule(module engine.ModuleInfo) error {
	return m.engine.Start(module)
}

func (m *Master) stopModule(module string) error {
	return m.engine.Stop(module)
}
*/

func defaults(c *Config) error {
	if runtime.GOOS == "linux" {
		err := os.MkdirAll("var/run", os.ModePerm)
		if err != nil {
			openedge.WithError(err).Errorf("failed to make dir: var/run")
		}
		c.Server = "unix://var/run/openedge.sock"
		utils.SetEnv(openedge.MasterAPIKey, c.Server)
	} else {
		if c.Server == "" {
			c.Server = "tcp://127.0.0.1:50050"
		}
		addr := c.Server
		uri, err := url.Parse(addr)
		if err != nil {
			return err
		}
		if c.Mode == "docker" {
			parts := strings.SplitN(uri.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
		}
		utils.SetEnv(openedge.MasterAPIKey, addr)
	}
	utils.SetEnv(openedge.HostOSKey, runtime.GOOS)
	utils.SetEnv(openedge.ModuleModeKey, c.Mode)
	return nil
}
