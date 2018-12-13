package http

import (
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/baidu/openedge/trans"
	"github.com/baidu/openedge/utils"
	"github.com/gorilla/mux"
)

// Params http params
type Params = map[string]string

// Headers http headers
type Headers = http.Header

// Server http server
type Server struct {
	Address string
	*http.Server
	uri    *url.URL
	router *mux.Router
}

// NewServer creates a new http server
func NewServer(sc ServerConfig) (*Server, error) {
	uri, err := utils.ParseURL(sc.Address)
	if err != nil {
		return nil, err
	}
	tls, err := trans.NewTLSServerConfig(sc.CA, sc.Key, sc.Cert)
	if err != nil {
		return nil, err
	}
	router := mux.NewRouter()
	return &Server{
		Server: &http.Server{
			WriteTimeout: sc.Timeout,
			ReadTimeout:  sc.Timeout,
			TLSConfig:    tls,
			Handler:      router,
		},
		router: router,
		uri:    uri,
	}, nil
}

// Handle handle requests
func (s *Server) Handle(handle func(Params, Headers, []byte) ([]byte, error), method, path string, params ...string) {
	s.router.HandleFunc(path, func(res http.ResponseWriter, req *http.Request) {
		var err error
		var reqBody []byte
		if req.Body != nil {
			reqBody, err = ioutil.ReadAll(req.Body)
			if err != nil {
				http.Error(res, err.Error(), 400)
				return
			}
		}
		resBody, err := handle(mux.Vars(req), req.Header, reqBody)
		if err != nil {
			http.Error(res, err.Error(), 400)
			return
		}
		if resBody != nil {
			res.Write(resBody)
		}
	}).Methods(method).Queries(params...)
}

// Start starts server
func (s *Server) Start() error {
	l, err := net.Listen(s.uri.Scheme, s.uri.Host)
	if err != nil {
		return err
	}
	if s.uri.Scheme == "tcp" {
		l = tcpKeepAliveListener{l.(*net.TCPListener)}
	}
	s.Addr = l.Addr().String()
	go s.Serve(l)
	return nil
}

// Close closese server
func (s *Server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.IdleTimeout)
	defer cancel()
	return s.Shutdown(ctx)
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
