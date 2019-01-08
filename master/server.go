package master

import (
	"encoding/json"
	"fmt"

	"github.com/baidu/openedge/module/config"
	"github.com/baidu/openedge/module/http"
	"github.com/baidu/openedge/module/logger"
	"github.com/baidu/openedge/module/master"
	"github.com/baidu/openedge/module/utils"
)

type api interface {
	stats() *master.Stats
	reload(file string) error
	authModule(username, password string) bool
	startModule(module config.Module) error
	stopModule(module string) error
}

// Server api server to start/stop modules
type Server struct {
	*http.Server
	api api
	log *logger.Entry
}

// NewServer creates a new server
func NewServer(a api, c config.HTTPServer) (*Server, error) {
	svr, err := http.NewServer(c)
	if err != nil {
		return nil, err
	}
	s := &Server{
		Server: svr,
		api:    a,
		log:    logger.WithFields("api", "http"),
	}
	s.Handle(s.stats, "GET", "/stats")
	s.Handle(s.reload, "PUT", "/reload", "file", "{file}")
	s.Handle(s.getPort, "GET", "/ports/available", "host", "{host}")
	s.Handle(s.startModule, "PUT", "/modules/{name}/start")
	s.Handle(s.stopModule, "PUT", "/modules/{name}/stop")
	return s, nil
}

func (s *Server) stats(params http.Params, headers http.Headers, reqBody []byte) ([]byte, error) {
	if !s.api.authModule(headers.Get("x-iot-edge-username"), headers.Get("x-iot-edge-password")) {
		account := headers.Get("x-iot-edge-username")
		s.log.Errorf("unauthorized to get port by account (%s)", account)
		return nil, fmt.Errorf("account (%s) unauthorized", account)
	}
	resBody, err := json.Marshal(s.api.stats())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (s *Server) reload(params http.Params, headers http.Headers, reqBody []byte) ([]byte, error) {
	if !s.api.authModule(headers.Get("x-iot-edge-username"), headers.Get("x-iot-edge-password")) {
		account := headers.Get("x-iot-edge-username")
		s.log.Errorf("unauthorized to get port by account (%s)", account)
		return nil, fmt.Errorf("account (%s) unauthorized", account)
	}
	file, ok := params["file"]
	if !ok {
		return nil, fmt.Errorf("param 'file' missing")
	}
	go s.api.reload(file)
	return nil, nil
}

func (s *Server) startModule(params http.Params, headers http.Headers, reqBody []byte) ([]byte, error) {
	if !s.api.authModule(headers.Get("x-iot-edge-username"), headers.Get("x-iot-edge-password")) {
		account := headers.Get("x-iot-edge-username")
		s.log.Errorf("unauthorized to start module (%s) by account (%s)", params["name"], account)
		return nil, fmt.Errorf("account (%s) unauthorized", account)
	}
	if reqBody == nil {
		return nil, fmt.Errorf("request body missing")
	}
	var m config.Module
	err := utils.UnmarshalJSON(reqBody, &m)
	if err != nil {
		return nil, err
	}
	if err = s.api.startModule(m); err != nil {
		s.log.WithError(err).Errorf("failed to start module (%s)", m.UniqueName())
		return nil, err
	}
	return nil, nil
}

func (s *Server) stopModule(params http.Params, headers http.Headers, reqBody []byte) ([]byte, error) {
	if !s.api.authModule(headers.Get("x-iot-edge-username"), headers.Get("x-iot-edge-password")) {
		account := headers.Get("x-iot-edge-username")
		s.log.Errorf("unauthorized to stop module (%s) by account (%s)", params["name"], account)
		return nil, fmt.Errorf("account (%s) unauthorized", account)
	}
	if err := s.api.stopModule(params["name"]); err != nil {
		s.log.WithError(err).Errorf("failed to stop module (%s)", params["name"])
		return nil, err
	}
	return nil, nil
}

func (s *Server) getPort(params http.Params, headers http.Headers, reqBody []byte) ([]byte, error) {
	if !s.api.authModule(headers.Get("x-iot-edge-username"), headers.Get("x-iot-edge-password")) {
		account := headers.Get("x-iot-edge-username")
		s.log.Errorf("unauthorized to get port by account (%s)", account)
		return nil, fmt.Errorf("account (%s) unauthorized", account)
	}
	host, ok := params["host"]
	if !ok {
		host = "127.0.0.1"
	}
	port, err := utils.GetPortAvailable(host)
	if err != nil {
		return nil, err
	}
	data := map[string]int{"port": port}
	resBody, err := json.Marshal(&data)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}
