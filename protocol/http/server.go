package http

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"syscall"
	"time"

	"github.com/baetyl/baetyl/logger"
	"github.com/baetyl/baetyl/utils"
	"github.com/creasty/defaults"
	"github.com/gorilla/mux"
)

// Params http params
type Params = map[string]string

// Headers http headers
type Headers = http.Header

// Server http server
type Server struct {
	cfg    ServerInfo
	svr    *http.Server
	uri    *url.URL
	addr   string
	auth   func(u, p string) bool
	router *mux.Router
	log    logger.Logger
}

// NewServer creates a new http server
func NewServer(c ServerInfo, a func(u, p string) bool) (*Server, error) {
	defaults.Set(&c)

	uri, err := utils.ParseURL(c.Address)
	if err != nil {
		return nil, err
	}
	tls, err := utils.NewTLSServerConfig(c.Certificate)
	if err != nil {
		return nil, err
	}
	router := mux.NewRouter()
	return &Server{
		cfg:    c,
		auth:   a,
		uri:    uri,
		router: router,
		svr: &http.Server{
			WriteTimeout: c.Timeout,
			ReadTimeout:  c.Timeout,
			TLSConfig:    tls,
			Handler:      router,
		},
		log: logger.WithField("api", "server"),
	}, nil
}

// Handle handle requests
func (s *Server) Handle(handle func(Params, []byte) ([]byte, error), method, path string, params ...string) {
	s.router.HandleFunc(path, func(res http.ResponseWriter, req *http.Request) {
		s.log.Infof("[%s] %s", req.Method, req.URL.String())
		if s.auth != nil {
			if !s.auth(req.Header.Get(headerKeyUsername), req.Header.Get(headerKeyPassword)) {
				http.Error(res, errAccountUnauthorized.Error(), 401)
				s.log.Errorf("[%s] %s %s", req.Method, req.URL.String(), errAccountUnauthorized.Error())
				return
			}
		}
		var err error
		var reqBody []byte
		if req.Body != nil {
			defer req.Body.Close()
			reqBody, err = ioutil.ReadAll(req.Body)
			if err != nil {
				http.Error(res, err.Error(), 400)
				s.log.Errorf("[%s] %s %s", req.Method, req.URL.String(), err.Error())
				return
			}
		}
		resBody, err := handle(mux.Vars(req), reqBody)
		if err != nil {
			http.Error(res, err.Error(), 400)
			s.log.Errorf("[%s] %s %s", req.Method, req.URL.String(), err.Error())
			return
		}
		if resBody != nil {
			res.Write(resBody)
		}
	}).Methods(method).Queries(params...)
}

// Start starts server
func (s *Server) Start() error {
	if s.uri.Scheme == "unix" {
		if err := syscall.Unlink(s.uri.Host); err != nil {
			s.log.Errorf(err.Error())
		}
	}

	l, err := net.Listen(s.uri.Scheme, s.uri.Host)
	if err != nil {
		return err
	}
	if s.uri.Scheme == "tcp" {
		l = tcpKeepAliveListener{l.(*net.TCPListener)}
	}
	s.addr = l.Addr().String()
	go s.svr.Serve(l)
	return nil
}

// Close closese server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.svr.IdleTimeout)
	defer cancel()
	return s.svr.Shutdown(ctx)
}

// tcpKeepAliveListener sets TCP keep-alive timeouts on accepted
// connections. It's used by ListenAndServe and ListenAndServeTLS so
// dead TCP connections (e.g. closing laptop mid-download) eventually
// go away.
type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}
