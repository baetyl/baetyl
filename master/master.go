package master

import (
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"runtime"
	"strings"

	openedge "github.com/baidu/openedge/api/go"
	"github.com/baidu/openedge/master/engine"
	sdk "github.com/baidu/openedge/sdk/go"
	"github.com/baidu/openedge/utils"
	cmap "github.com/orcaman/concurrent-map"
)

// Version of master
var Version string

// Master master manages all modules and connects with cloud
type Master struct {
	cfg      Config
	dyncfg   DynamicConfig
	engine   engine.Engine
	server   *Server
	services cmap.ConcurrentMap
	workdir  string
	log      openedge.Logger
}

// New creates a new master
func New(confpath string) (*Master, error) {
	wdir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(path.Join(wdir, confpath))
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = utils.UnmarshalYAML(data, &cfg)
	if err != nil {
		return nil, err
	}
	err = defaults(&cfg)
	if err != nil {
		return nil, err
	}
	err = sdk.InitLogger(&cfg.Logger, "openedge", "master")
	if err != nil {
		return nil, err
	}
	openedge.Debugln("work dir:", wdir)

	m := &Master{
		cfg:      cfg,
		workdir:  wdir,
		services: cmap.New(),
		log:      openedge.WithField("openedge", "master"),
	}
	m.engine, err = engine.New(cfg.Mode, wdir)
	if err != nil {
		m.Close()
		return nil, err
	}
	m.server, err = newServer(m)
	if err != nil {
		m.Close()
		return nil, err
	}
	err = m.load()
	if err != nil {
		m.Close()
		return nil, err
	}
	err = m.startServices()
	if err != nil {
		m.Close()
		return nil, err
	}
	return m, nil
}

// Close closes agent
func (m *Master) Close() error {
	m.cleanServices()
	if m.server != nil {
		m.server.close()
	}
	if m.server != nil {
		m.engine.Close()
	}
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
