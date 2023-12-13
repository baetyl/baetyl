// Package ws 实现端云基于ws协议的链接
package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	v2plugin "github.com/baetyl/baetyl-go/v2/plugin"
	specV1 "github.com/baetyl/baetyl-go/v2/spec/v1"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/gorilla/websocket"
	"github.com/jpillora/backoff"

	"github.com/baetyl/baetyl/v2/common"
	v2initz "github.com/baetyl/baetyl/v2/initz"
	"github.com/baetyl/baetyl/v2/plugin"
)

var (
	ErrLinkTLSConfigMissing = errors.New("certificate bidirectional authentication is required for connection with cloud")
	ErrConnectNotRunning    = errors.New("websocket has no available link and cannot send data")
)

type wsLink struct {
	mutex       sync.Mutex
	dialer      websocket.Dialer
	urls        []string
	keeper      common.SendKeeper
	cfg         Config
	reconnectCh chan struct{}
	reNotifyCh  chan struct{}
	stateMutex  sync.RWMutex
	msgCh       chan *specV1.Message
	errCh       chan error
	state       *specV1.Message
	conn        *websocket.Conn
	isConnRun   atomic.Value
	ctx         context.Context
	cancel      context.CancelFunc
	log         *log.Logger
	backoff     backoff.Backoff
}

func (ws *wsLink) Close() error {
	ws.cancel()
	ws.mutex.Lock()
	defer ws.mutex.Unlock()
	if ws.conn != nil {
		return ws.conn.Close()
	}
	return nil
}

func init() {
	v2plugin.RegisterFactory("wslink", New)
}

func New() (v2plugin.Plugin, error) {
	var cfg Config
	if err := utils.LoadYAML(plugin.ConfFile, &cfg); err != nil {
		return nil, errors.Trace(err)
	}
	log.L().Debug("config", log.Any("cfg", cfg))

	cfg.WSLink.Certificate.CA = cfg.Node.CA
	cfg.WSLink.Certificate.Key = cfg.Node.Key
	cfg.WSLink.Certificate.Cert = cfg.Node.Cert
	ops, err := cfg.WSLink.ToClientOptions()
	if err != nil {
		return nil, errors.Trace(err)
	}
	if ops.TLSConfig == nil {
		return nil, errors.Trace(ErrLinkTLSConfigMissing)
	}

	dialer := websocket.Dialer{
		NetDial: nil,
		NetDialContext: (&net.Dialer{
			Timeout:   ops.Timeout,
			KeepAlive: ops.KeepAlive,
		}).DialContext,
		Proxy:            http.ProxyFromEnvironment,
		TLSClientConfig:  ops.TLSConfig,
		HandshakeTimeout: ops.TLSHandshakeTimeout,
	}

	addrs := strings.Split(cfg.WSLink.Address, ",")
	urls := []string{}
	for _, addr := range addrs {
		urls = append(urls, fmt.Sprintf("%s/%s", addr, cfg.WSLink.SyncURL))
	}

	ctx, cancel := context.WithCancel(context.Background())

	link := &wsLink{
		dialer:      dialer,
		keeper:      common.SendKeeper{},
		urls:        urls,
		isConnRun:   atomic.Value{},
		reconnectCh: make(chan struct{}, 1),
		reNotifyCh:  make(chan struct{}, 1),
		msgCh:       make(chan *specV1.Message, 1),
		errCh:       make(chan error, 1),
		state:       &specV1.Message{Kind: plugin.LinkStateUnknown, Content: specV1.LazyValue{Value: ""}},
		cfg:         cfg,
		ctx:         ctx,
		cancel:      cancel,
		backoff:     backoff.Backoff{Min: cfg.WSLink.ReconnectBackoff.Min, Max: cfg.WSLink.ReconnectBackoff.Max, Factor: cfg.WSLink.ReconnectBackoff.Factor},
		log:         log.With(log.Any("plugin", "wslink")),
	}
	if addrEnv := os.Getenv(v2initz.KeyBaetylSyncAddr); addrEnv != "" {
		addrs = strings.Split(addrEnv, ",")
		for _, addr := range addrs {
			urls = append(urls, fmt.Sprintf("%s/%s", addr, cfg.WSLink.SyncURL))
		}
		link.urls = urls
	}
	link.dial()
	go link.receiving()
	go link.reconnecting()
	return link, nil
}

func (ws *wsLink) dial() {
	for _, url := range ws.urls {
		conn, _, err := ws.dialer.DialContext(ws.ctx, url, nil)
		if err != nil {
			ws.log.Warn("failed to connect cloud", log.Any("url", url), log.Error(err))
		} else {
			ws.conn = conn
			ws.isConnRun.Store(true)
			return
		}
	}
	ws.isConnRun.Store(false)
}

func (ws *wsLink) Send(msg *specV1.Message) error {
	flag := ws.isConnRun.Load()
	if flag != nil {
		if isRun, ok := flag.(bool); ok && !isRun {
			ws.log.Debug("connect is reconnecting")
			return errors.Trace(ErrConnectNotRunning)
		}
	}
	ws.mutex.Lock()
	err := ws.conn.WriteJSON(msg)
	ws.mutex.Unlock()
	if err != nil {
		select {
		case ws.reconnectCh <- struct{}{}:
		case <-ws.ctx.Done():
			return nil
		default:
		}
		ws.log.Error("failed to send message", log.Error(err))
		return errors.Trace(err)
	}
	return nil
}

func (ws *wsLink) receiving() {
	for {
		var (
			r   io.Reader
			err error
		)
		if ws.conn != nil {
			err = ws.conn.SetReadDeadline(time.Now().Add(ws.cfg.Sync.Report.Interval + ws.cfg.WSLink.WaitResponseInterval))
			if err != nil {
				ws.log.Warn("failed to set read timeout", log.Error(err))
			}
			_, r, err = ws.conn.NextReader()
		}
		if err != nil || ws.conn == nil {
			ws.log.Warn("failed to get websocket reader, websocket will reconnect", log.Error(err))
			select {
			case ws.reconnectCh <- struct{}{}:
			case <-ws.ctx.Done():
				return
			default:
			}
			select {
			case <-ws.reNotifyCh:
				ws.log.Debug("websocket reconnect successfully, connect will get next reader")
				continue
			case <-ws.ctx.Done():
				return
			}
		}
		msg := new(specV1.Message)
		err = json.NewDecoder(r).Decode(msg)
		if err != nil {
			ws.log.Debug("failed tod decode message", log.Error(err))
			continue
		}

		data, err := utils.ParseEnv(msg.Content.GetJSON())
		if err != nil {
			ws.log.Error("failed to parse env", log.Error(err))
			continue
		}
		msg.Content.SetJSON(data)
		if common.IsSyncMessage(msg) {
			err = ws.keeper.ReceiveResp(msg)
			if err != nil {
				ws.log.Error("failed to receive response", log.Error(err))
				continue
			}
		} else {
			if msg.Kind == specV1.MessageError {
				var errMsg string
				err = msg.Content.Unmarshal(&errMsg)
				if err != nil {
					ws.log.Error("failed to unmarshal error message", log.Error(err))
					continue
				}
				ws.log.Debug("get cloud error message", log.Any("errMsg", errMsg))
				select {
				case ws.errCh <- errors.New(errMsg):
				case <-ws.ctx.Done():
					return
				}
				if strings.HasSuffix(errMsg, fmt.Sprintf("The (node) resource (%s) is not found.", msg.Metadata["name"])) {
					ws.stateNotify(plugin.LinkStateNodeNotFound, errMsg)
				} else {
					ws.stateNotify(plugin.LinkStateSucceeded, errMsg)
				}
				continue
			}
			ws.stateNotify(plugin.LinkStateSucceeded, plugin.LinkStateSucceeded)
			select {
			case ws.msgCh <- msg:
			case <-ws.ctx.Done():
				return
			}
		}
	}
}

// State RLock and return the copy of state pointer
func (ws *wsLink) State() *specV1.Message {
	ws.stateMutex.RLock()
	copyState := ws.state
	ws.stateMutex.RUnlock()
	return copyState
}

func (ws *wsLink) Receive() (<-chan *specV1.Message, <-chan error) {
	return ws.msgCh, ws.errCh
}

func (ws *wsLink) IsAsyncSupported() bool {
	return true
}

func (ws *wsLink) Request(msg *specV1.Message) (*specV1.Message, error) {
	res, err := ws.keeper.SendSync(msg, ws.cfg.WSLink.Timeout, ws.Send)
	if err != nil {
		return nil, errors.Trace(err)
	}
	// encapsulation error message
	if res.Kind == specV1.MessageError {
		var errMsg string
		err = res.Content.Unmarshal(&errMsg)
		if err != nil {
			return nil, errors.Trace(err)
		}
		return nil, errors.New(errMsg)
	}
	return res, nil
}

func (ws *wsLink) reconnecting() {
	for {
		select {
		case <-ws.reconnectCh:
			ws.isConnRun.Store(false)
			if ws.conn != nil {
				ws.mutex.Lock()
				err := ws.conn.Close()
				if err != nil {
					ws.log.Debug("failed to close ws connect", log.Any("conn", ws.conn), log.Error(err))
				}
				ws.conn = nil
				ws.mutex.Unlock()
			}
			time.Sleep(ws.backoff.Duration())

			errs := []string{}
			var conn *websocket.Conn
			var err error
			for _, url := range ws.urls {
				conn, _, err = ws.dialer.DialContext(ws.ctx, url, nil)
				if err != nil {
					ws.log.Warn("failed to connect cloud", log.Any("url", url), log.Error(err))
					errs = append(errs, err.Error())
					continue
				}
				break
			}

			if len(errs) == len(ws.urls) {
				ws.log.Error("failed to reconnect websocket", log.Any("errors", strings.Join(errs, ";")))
				ws.stateNotify(plugin.LinkStateNetworkError, strings.Join(errs, ";"))
				select {
				case ws.reconnectCh <- struct{}{}:
				case <-ws.ctx.Done():
					return
				default:
				}
				continue
			}

			ws.mutex.Lock()
			ws.conn = conn
			ws.mutex.Unlock()
			ws.log.Info("websocket reconnected")
			ws.backoff.Reset()

			select {
			case ws.reNotifyCh <- struct{}{}:
			case <-ws.ctx.Done():
				return
			}
			ws.isConnRun.Store(true)
		case <-ws.ctx.Done():
			return
		}
	}
}

// stateNotify Lock and update the pointer of state
func (ws *wsLink) stateNotify(kind, msg string) {
	ws.stateMutex.Lock()
	ws.state = &specV1.Message{
		Kind:     specV1.MessageKind(kind),
		Metadata: map[string]string{},
		Content:  specV1.LazyValue{Value: msg},
	}
	ws.stateMutex.Unlock()
}
