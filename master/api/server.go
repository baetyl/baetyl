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

	// for system
	InspectSystem() *openedge.Inspect
	UpdateSystem(string, bool) error

	// for instance
	ReportInstance(serviceName, instanceName string, stats map[string]interface{}) error
	StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
	StopInstance(serviceName, instanceName string) error
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
	s.s.Handle(s.reportInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/report")
	s.s.Handle(s.startInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/start")
	s.s.Handle(s.stopInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/stop")
	return s, s.s.Start()
}

// Close closes api server
func (s *Server) Close() error {
	return s.s.Close()
}

func (s *Server) inspectSystem(_ http.Params, reqBody []byte) ([]byte, error) {
	return json.Marshal(s.m.InspectSystem())
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
	return json.Marshal(res)
}

func (s *Server) reportInstance(params http.Params, reqBody []byte) ([]byte, error) {
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
	stats := make(map[string]interface{})
	err := json.Unmarshal(reqBody, &stats)
	if err != nil {
		return nil, err
	}
	return nil, s.m.ReportInstance(serviceName, instanceName, stats)
}

func (s *Server) startInstance(params http.Params, reqBody []byte) ([]byte, error) {
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
	return nil, s.m.StartInstance(serviceName, instanceName, dynamicConfig)
}

func (s *Server) stopInstance(params http.Params, _ []byte) ([]byte, error) {
	serviceName, ok := params["serviceName"]
	if !ok {
		return nil, fmt.Errorf("request params invalid, missing service name")
	}
	instanceName, ok := params["instanceName"]
	if !ok {
		return nil, fmt.Errorf("request params invalid, missing instance name")
	}
	return nil, s.m.StopInstance(serviceName, instanceName)
}
