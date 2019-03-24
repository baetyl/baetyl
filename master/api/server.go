package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/baidu/openedge/protocol/http"
	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
)

// Master master interface
type Master interface {
	Auth(u, p string) bool
	InspectSystem() *openedge.Inspect
	UpdateSystem(string, bool) error

	StartServiceInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
	StopServiceInstance(serviceName, instanceName string) error
}

// Server master api server
type Server struct {
	m Master
	s *http.Server
}

// New creates new api server
func New(c http.ServerInfo, m Master) (*Server, error) {
	svr, err := http.NewServer(c, m.Auth)
	if err != nil {
		return nil, err
	}
	s := &Server{
		m: m,
		s: svr,
	}
	s.s.Handle(s.inspectSystem, "GET", "/system/inspect")
	s.s.Handle(s.updateSystem, "PUT", "/system/update")

	s.s.Handle(s.getAvailablePort, "GET", "/ports/available")
	s.s.Handle(s.startServiceInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/start")
	s.s.Handle(s.stopServiceInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/stop")
	return s, s.s.Start()
}

// Close closes api server
func (s *Server) Close() error {
	return s.s.Close()
}

func (s *Server) inspectSystem(_ http.Params, reqBody []byte) ([]byte, error) {
	resBody, err := json.Marshal(s.m.InspectSystem())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (s *Server) updateSystem(_ http.Params, reqBody []byte) ([]byte, error) {
	if reqBody == nil {
		return nil, fmt.Errorf("request body invalid")
	}
	args := make(map[string]string)
	err := json.Unmarshal(reqBody, &args)
	if err != nil {
		return nil, err
	}
	clean := false
	if s, ok := args["clean"]; ok && strings.ToLower(s) == "true" {
		clean = true
	}
	go s.m.UpdateSystem(args["file"], clean)
	return nil, nil
}

func (s *Server) getAvailablePort(_ http.Params, reqBody []byte) ([]byte, error) {
	port, err := utils.GetAvailablePort("127.0.0.1")
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	res["port"] = strconv.Itoa(port)
	resBody, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (s *Server) startServiceInstance(params http.Params, reqBody []byte) ([]byte, error) {
	if reqBody == nil {
		return nil, fmt.Errorf("request body invalid")
	}
	serviceName, ok := params["serviceName"]
	if !ok {
		return nil, fmt.Errorf("request params invalid, missing service name")
	}
	instanceName, ok := params["instanceName"]
	if !ok {
		return nil, fmt.Errorf("request params invalid, missing instance name")
	}
	dynamicConfig := make(map[string]string)
	err := json.Unmarshal(reqBody, &dynamicConfig)
	if err != nil {
		return nil, err
	}
	err = s.m.StartServiceInstance(serviceName, instanceName, dynamicConfig)
	return nil, err
}

func (s *Server) stopServiceInstance(params http.Params, _ []byte) ([]byte, error) {
	serviceName, ok := params["serviceName"]
	if !ok {
		return nil, fmt.Errorf("request params invalid, missing service name")
	}
	instanceName, ok := params["instanceName"]
	if !ok {
		return nil, fmt.Errorf("request params invalid, missing instance name")
	}
	err := s.m.StopServiceInstance(serviceName, instanceName)
	return nil, err
}
