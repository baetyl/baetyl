package baetyl

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/schema/v3"
	"github.com/baetyl/baetyl/utils"

	cmap "github.com/orcaman/concurrent-map"
	"google.golang.org/grpc"
	"gopkg.in/tomb.v2"
)

var Version string
var Revision string
var DefaultPrefix string
var DefaultConfigPath string
var DefaultLoggerPath string
var DefaultDataPath string
var DefaultPidPath string
var DefaultAPIAddress string

type runtime struct {
	cfg     Config
	log     logger.Logger
	e       Engine
	db      Database
	cert    certificate
	license license
	stats
	srv      *grpc.Server
	lsrv     legacyServer
	services cmap.ConcurrentMap
	accounts cmap.ConcurrentMap
	atask    task
	utask    task
}

type Context interface {
	Logger() logger.Logger
	Config() *Config
	UpdateStats(svc, inst string, stats schema.InstanceStats)
	RemoveServiceStats(svc string)
	RemoveInstanceStats(svc, inst string)
}

func (rt *runtime) Logger() logger.Logger {
	return rt.log
}

func (rt *runtime) Config() *Config {
	return &rt.cfg
}

// Run a unique instance within the given context
func Run(prefix, configPath string) error {
	rt := runtime{
		stats: newStats(
			defaultEngine,
			Version,
			Revision,
		),
		services: cmap.New(),
		accounts: cmap.New(),
	}
	t, ctx := tomb.WithContext(context.TODO())
	t.Go(func() error { return rt.run(ctx, prefix, configPath) })
	signal.Ignore(syscall.SIGPIPE)
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
	select {
	case <-sig:
		t.Kill(nil)
	case <-t.Dead():
	}
	return t.Wait()
}

func (rt *runtime) run(ctx context.Context, prefix, configPath string) error {
	var err error

	rt.atask.c = ctx
	rt.utask.c = ctx

	err = rt.initSystem(prefix, configPath)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(rt.cfg.PidPath, []byte(fmt.Sprintf("%d", os.Getpid())), 0644)
	if err != nil {
		return err
	}
	defer os.Remove(rt.cfg.PidPath)

	err = rt.initLogger()
	if err != nil {
		return err
	}

	err = rt.startEngine()
	if err != nil {
		return err
	}
	defer rt.stopEngine()

	err = rt.openDB()
	if err != nil {
		return err
	}
	defer rt.closeDB()

	err = rt.setupSysCert()
	if err != nil {
		return err
	}

	err = rt.loadLicense()
	if err != nil {
		return err
	}

	t1, ctx1 := tomb.WithContext(ctx)
	t1.Go(func() error { return rt.runServer(ctx1) })
	t1.Go(func() error { return rt.runLegacyServer(ctx1) })

	rt.runLocal(ctx)
	rt.localUpdate(ctx)

	t2, ctx2 := tomb.WithContext(ctx)
	t2.Go(func() error { return rt.runReport(ctx2) })
	t2.Go(func() error { return rt.runDesire(ctx2) })

	<-ctx.Done()
	t2.Wait()
	t1.Wait()
	return nil
}

func (rt *runtime) initSystem(prefix, configPath string) error {
	if prefix == "" {
		prefix = DefaultPrefix
	}
	if !filepath.IsAbs(prefix) {
		val, err := filepath.Abs(prefix)
		if err != nil {
			return err
		}
		prefix = val
	}
	if configPath == "" {
		configPath = DefaultConfigPath
	}
	if !filepath.IsAbs(configPath) {
		configPath = filepath.Join(prefix, configPath)
	}
	data, err := ioutil.ReadFile(configPath)
	if err == nil {
		res, err := utils.ParseEnv(data)
		if err != nil {
			fmt.Fprintf(os.Stderr, "config parse error: %s", err.Error())
			res = data
		}
		err = utils.UnmarshalYAML(res, &rt.cfg)
	} else if os.IsNotExist(err) {
		err = utils.SetDefaults(&rt.cfg)
	}
	if err != nil {
		return err
	}
	return rt.cfg.validate(prefix)
}

func (rt *runtime) initLogger() error {
	rt.log = logger.InitLogger(rt.cfg.Logger, "baetyl", "master")
	cfg := &rt.cfg
	rt.log.Infof(
		"grace: %d; data: %s; apiserver: %s://%s; legacy_apiserver: %s://%s",
		cfg.Grace,
		cfg.DataPath,
		cfg.APIConfig.Network, cfg.APIConfig.Address,
		cfg.LegacyAPIConfig.Network, cfg.LegacyAPIConfig.Address,
	)
	return nil
}

func (rt *runtime) runLocal(ctx context.Context) {
	ac, err := rt.loadAppConfig(rt.cfg.DataPath)
	if err != nil {
		rt.log.Warnf("load local app config fail: %s", err.Error())
		return
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	rt.atask.run(func(ctx context.Context) error {
		return rt.apply(ctx, ac)
	})
}

// apply ComposeAppConfig to engine with injected resources
// NOT thread safe, SHOULD be protected by task
func (rt *runtime) apply(ctx context.Context, ac *schema.ComposeAppConfig) error {
	rt.log.Infof("begin apply application [version=%s]", ac.AppVersion)
	defer rt.log.Infof("complete apply application [version=%s]", ac.AppVersion)
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	err := rt.injectAppConfig(ctx, ac)
	if err != nil {
		rt.log.Warnf("inject appconfig fail: %s", err.Error())
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	return rt.e.Apply(ctx, ac)
}
