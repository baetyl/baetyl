package master

import (
	"fmt"
	"net/url"
	"os"
	"path"
	"regexp"
	"runtime"
	"strings"

	"github.com/baidu/openedge/agent"
	"github.com/baidu/openedge/api"
	"github.com/baidu/openedge/config"
	"github.com/baidu/openedge/engine"
	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/module"
	"github.com/baidu/openedge/trans/mqtt"
	"github.com/juju/errors"
	"github.com/mholt/archiver"
)

// dirs to backup
const appDir = "app"
const appBackupFile = "app.bk"

// app config file
var appConfFile = path.Join(appDir, "app.yml")

// Master master manages all modules and connects with cloud
type Master struct {
	conf   config.Master
	engine *engine.Engine
	agent  *agent.Agent
	api    *api.Server
	pwd    string
}

// New creates a new master
func New(confDate string) (*Master, error) {
	c := config.Master{}
	err := module.Load(&c, confDate)
	if err != nil {
		return nil, errors.Trace(err)
	}
	err = defaults(&c)
	if err != nil {
		return nil, errors.Trace(err)
	}
	logger.Init(c.Logger, "openedge", "master")
	ctx := engine.Context{
		Mode:  c.Mode,
		Grace: c.Grace,
	}
	en, err := engine.New(&ctx)
	if err != nil {
		return nil, errors.Trace(err)
	}
	ap, err := api.NewServer(en, c.API)
	if err != nil {
		return nil, errors.Trace(err)
	}

	var ag *agent.Agent
	if c.Cloud.Address != "" {
		ag, err = agent.NewAgent(c.Cloud)
		if err != nil {
			ap.Close()
			return nil, errors.Trace(err)
		}
	}

	pwd, err := os.Getwd()
	if err != nil {
		return nil, errors.Trace(err)
	}
	m := &Master{
		conf:   c,
		engine: en,
		agent:  ag,
		api:    ap,
		pwd:    pwd,
	}
	err = m.api.Start()
	if err != nil {
		m.Close()
		return nil, errors.Trace(err)
	}
	if m.agent != nil {
		if err := m.agent.Start(m.Reload); err != nil {
			m.Close()
			return nil, errors.Trace(err)
		}
	}
	return m, nil
}

// Start starts agent
func (m *Master) Start() error {
	err := m.loadAppConfig()
	if err != nil {
		return errors.Trace(err)
	}
	if err := m.engine.StartAll(m.conf.Modules); err != nil {
		return errors.Trace(err)
	}
	if m.agent != nil {
		report := map[string]interface{}{
			"mode":         m.conf.Mode,
			"conf_version": m.conf.Version,
		}
		m.agent.Report(report)
	}
	return nil
}

// Close closes agent
func (m *Master) Close() {
	if m.agent != nil {
		if err := m.agent.Close(); err != nil {
			logger.WithError(err).Errorf("Failed to close cloud agent")
		}
	}
	m.engine.StopAll()
	if err := m.api.Close(); err != nil {
		logger.WithError(err).Errorf("Failed to close api server")
	}
}

// Reload reload app
func (m *Master) Reload(version string) map[string]interface{} {
	err := m.reload(version)
	report := map[string]interface{}{
		"mode":         m.conf.Mode,
		"conf_version": m.conf.Version,
	}
	if err != nil {
		report["reload_error"] = err.Error()
		logger.WithError(err).Error("Failed to reload app config")
	} else {
		logger.Infof("Loaded app config (version:%s)", m.conf.Version)
	}
	return report
}

func (m *Master) reload(version string) error {
	if !isVersion(version) {
		return errors.Errorf("New config version invalid")
	}
	err := m.backupAppDir()
	if err != nil {
		return errors.Errorf("Failed to backup old config: %s", err.Error())
	}
	defer m.cleanBackupFile()
	err = m.unpackConfigFile(version)
	if err != nil {
		return errors.Errorf("Failed to unpack new config: %s", err.Error())
	}
	err = m.loadAppConfig()
	if err != nil {
		return errors.Errorf("Failed to load new config: %s", err.Error())
	}
	m.engine.StopAll()
	err = m.engine.StartAll(m.conf.Modules)
	if err != nil {
		logger.WithError(err).Info("Failed to load new config, rollback")
		err1 := m.unpackBackupFile()
		if err1 != nil {
			err = errors.Errorf(err.Error() + ";Failed to unpack old config backup file" + err1.Error())
			return errors.Trace(err)
		}
		err1 = m.loadAppConfig()
		if err1 != nil {
			err = errors.Errorf(err.Error() + ";Failed to load old config" + err1.Error())
			return errors.Trace(err)
		}
		m.engine.StopAll()
		err1 = m.engine.StartAll(m.conf.Modules)
		if err1 != nil {
			err = errors.Errorf(err.Error() + ";Failed to start modules with old config" + err.Error())
			return errors.Trace(err)
		}
	}
	return nil
}

func (m *Master) backupAppDir() error {
	if !dirExists(appDir) {
		os.MkdirAll(appDir, 0700)
	}
	return errors.Trace(archiver.Zip.Make(appBackupFile, []string{appDir}))
}

func (m *Master) cleanBackupFile() error {
	return errors.Trace(os.Remove(appBackupFile))
}

func (m *Master) unpackConfigFile(version string) error {
	file := version + ".zip"
	if !fileExists(file) {
		return errors.Errorf("app config zip file (%s) not found", file)
	}
	err := archiver.Zip.Open(file, m.pwd)
	return errors.Trace(err)
}

func (m *Master) unpackBackupFile() error {
	err := archiver.Zip.Open(appBackupFile, m.pwd)
	return errors.Trace(err)
}

func (m *Master) loadAppConfig() error {
	if !fileExists(appConfFile) {
		m.conf.Modules = []config.Module{}
		return nil
	}

	return errors.Trace(module.Load(&m.conf, appConfFile))
}

// IsVersion checks version
func isVersion(v string) bool {
	r := regexp.MustCompile("^[\\w\\.]+$")
	return r.MatchString(v)
}

// DirExists checkes file exists
func dirExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return fi.IsDir()
}

func fileExists(path string) bool {
	fi, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return !fi.IsDir()
}

func defaults(c *config.Master) error {
	if c.Cloud.Address != "" {
		backward := mqtt.Subscription{QOS: 1, Topic: fmt.Sprintf(agent.CloudBackward, c.Cloud.ClientID)}
		c.Cloud.Subscriptions = append(c.Cloud.Subscriptions, backward)
		if c.Cloud.OpenAPI.Address == "" {
			if strings.Contains(c.Cloud.Address, "bj.baidubce.com") {
				c.Cloud.OpenAPI.Address = "https://iotedge.bj.baidubce.com"
			} else if strings.Contains(c.Cloud.Address, "gz.baidubce.com") {
				c.Cloud.OpenAPI.Address = "https://iotedge.gz.baidubce.com"
			} else {
				return errors.Errorf("Cloud address invalid")
			}
		}
		if c.Cloud.OpenAPI.CA == "" {
			c.Cloud.OpenAPI.CA = "conf/openapi.pem"
		}
	}
	if runtime.GOOS == "linux" {
		c.API.Address = "unix://var/openedge.sock"
		module.SetEnv(module.EnvOpenEdgeMasterAPI, c.API.Address)
	} else {
		if c.API.Address == "" {
			c.API.Address = "tcp://127.0.0.1:50050"
		}
		addr := c.API.Address
		uri, err := url.Parse(addr)
		if err != nil {
			return errors.Trace(err)
		}
		if c.Mode == "docker" {
			parts := strings.SplitN(uri.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
		}
		module.SetEnv(module.EnvOpenEdgeMasterAPI, addr)
	}
	module.SetEnv(module.EnvOpenEdgeHostOS, runtime.GOOS)
	module.SetEnv(module.EnvOpenEdgeModuleMode, c.Mode)
	return nil
}
