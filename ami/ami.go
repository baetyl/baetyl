package ami

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	specv1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/gorilla/websocket"

	"github.com/baetyl/baetyl/v2/config"
	"github.com/baetyl/baetyl/v2/utils"
)

//go:generate mockgen -destination=../mock/ami.go -package=mock -source=ami.go AMI

const (
	BaetylGPUStatsExtension  = "baetyl_gpu_stats_extension"
	BaetylNodeStatsExtension = "baetyl_node_stats_extension"
	BaetylQPSStatsExtension  = "baetyl_qps_stats_extension"
)

var mu sync.Mutex
var amiNews = map[string]New{}
var amiImpls = map[string]AMI{}
var Hooks = map[string]interface{}{}

type CollectStatsExtFunc func(mode string) (map[string]interface{}, error)

type New func(cfg config.AmiConfig) (AMI, error)

type Pipe struct {
	InReader  *io.PipeReader
	InWriter  *io.PipeWriter
	OutReader *io.PipeReader
	OutWriter *io.PipeWriter
	Ctx       context.Context
	Cancel    context.CancelFunc
}

// AMI app model interfaces
type AMI interface {
	// node
	CollectNodeInfo() (map[string]interface{}, error)
	CollectNodeStats() (map[string]interface{}, error)
	GetModeInfo() (interface{}, error)

	// app
	ApplyApp(string, specv1.Application, map[string]specv1.Configuration, map[string]specv1.Secret) error
	DeleteApp(string, string) error
	StatsApps(string) ([]specv1.AppStats, error)

	// TODO: update
	FetchLog(namespace, service string, tailLines, sinceSeconds int64) (io.ReadCloser, error)

	// RemoteCommand remote debug
	RemoteCommand(option *DebugOptions, pipe Pipe) error

	// RemoteWebsocket remote link
	RemoteWebsocket(ctx context.Context, option *DebugOptions, pipe Pipe) error

	// RemoteLogs remote logs
	RemoteLogs(option *LogsOptions, pipe Pipe) error

	// RemoteDescribePod remote describe pod
	RemoteDescribe(tp, ns, n string) (string, error)

	UpdateNodeLabels(string, map[string]string) error

	// RPCApp call baetyl app from baetyl-core
	RPCApp(url string, req *specv1.RPCRequest) (*specv1.RPCResponse, error)
}

type DebugOptions struct {
	KubeDebugOptions
	NativeDebugOptions
	WebsocketOptions
}

type WebsocketOptions struct {
	Host string
	Path string
}

type KubeDebugOptions struct {
	Namespace string
	Name      string
	Container string
	Command   []string
}

type NativeDebugOptions struct {
	IP       string
	Port     string
	Username string
	Password string
}

type LogsOptions struct {
	Namespace    string
	Name         string
	Container    string
	SinceSeconds *int64
	TailLines    *int64
	LimitBytes   *int64
	Follow       bool
	Previous     bool
	Timestamps   bool
}

type ProcessInfo struct {
	Pid  int32
	Name string
	Ppid int32
}

func NewAMI(mode string, cfg config.AmiConfig) (AMI, error) {
	mu.Lock()
	defer mu.Unlock()
	if ami, ok := amiImpls[mode]; ok {
		return ami, nil
	}
	amiNew, ok := amiNews[mode]
	if !ok {
		return nil, errors.Trace(os.ErrInvalid)
	}
	ami, err := amiNew(cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	amiImpls[mode] = ami
	return ami, nil
}

func Register(name string, n New) {
	mu.Lock()
	defer mu.Unlock()
	if _, ok := amiNews[name]; ok {
		log.L().Warn("ami generator already exists, skip", log.Any("generator", name))
		return
	}
	log.L().Debug("ami generator registered", log.Any("generator", name))
	amiNews[name] = n
}

// RemoteWebsocket set up websocket connect
func RemoteWebsocket(ctx context.Context, option *DebugOptions, pipe Pipe) error {
	u := url.URL{Scheme: "ws", Host: option.WebsocketOptions.Host, Path: option.WebsocketOptions.Path}
	header := http.Header{
		"Sec-WebSocket-Protocol": []string{"web-proto"},
	}
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		return errors.Trace(err)
	}
	defer func() {
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.Close()
	}()
	// Check if the link to cloud is closed
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				pipe.OutReader.Close()
				pipe.InReader.Close()
				return
			case <-ticker.C:
			}
		}
	}()
	// Read from websocket and send to pipe
	go func() {
		for {
			msg, r, readErr := c.ReadMessage()
			if readErr != nil {
				log.L().Warn("failed to read remote message", log.Error(readErr))
				return
			}
			if msg == websocket.CloseMessage {
				return
			}
			_, readErr = pipe.OutWriter.Write(r)
			if readErr != nil {
				log.L().Error("failed to write up message", log.Error(readErr))
				return
			}
		}
	}()
	// Read from pipe util this reader is closed
	for {
		dt := make([]byte, utils.ReadBuff)
		n, writeErr := pipe.InReader.Read(dt)
		if writeErr != nil {
			log.L().Warn("InReader close", log.Error(writeErr))
			return nil
		}
		writeErr = c.WriteMessage(websocket.BinaryMessage, dt[0:n])
		if writeErr != nil {
			log.L().Error("failed to write remote message", log.Error(writeErr))
			continue
		}
	}
}
