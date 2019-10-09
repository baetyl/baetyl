package api

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/baetyl/baetyl/master/engine"
	"github.com/baetyl/baetyl/protocol/http"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

// Master master interface
type Master interface {
	Auth(u, p string) bool

	// for system
	InspectSystem() ([]byte, error)
	UpdateSystem(trace, tp, target string) error

	// for instance
	ReportInstance(serviceName, instanceName string, partialStats engine.PartialStats) error
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
	// v0, deprecated
	s.s.Handle(s.inspectSystemV0, "GET", "/system/inspect")
	s.s.Handle(s.updateSystem, "PUT", "/system/update")
	s.s.Handle(s.getAvailablePort, "GET", "/ports/available")
	s.s.Handle(s.startInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/start")
	s.s.Handle(s.stopInstance, "PUT", "/services/{serviceName}/instances/{instanceName}/stop")

	// v1
	s.s.Handle(s.inspectSystem, "GET", "/v1/system/inspect")
	s.s.Handle(s.updateSystem, "PUT", "/v1/system/update")
	s.s.Handle(s.getAvailablePort, "GET", "/v1/ports/available")
	s.s.Handle(s.reportInstance, "PUT", "/v1/services/{serviceName}/instances/{instanceName}/report")
	s.s.Handle(s.startInstance, "PUT", "/v1/services/{serviceName}/instances/{instanceName}/start")
	s.s.Handle(s.stopInstance, "PUT", "/v1/services/{serviceName}/instances/{instanceName}/stop")
	return s, s.s.Start()
}

// Close closes api server
func (s *Server) Close() error {
	return s.s.Close()
}

func (s *Server) inspectSystem(_ http.Params, reqBody []byte) ([]byte, error) {
	return s.m.InspectSystem()
}

/**********************************
agent version < 0.1.4
{
	"file": "var/db/baetyl/app/V1"
}
***********************************/
/**********************************
agent version = 0.1.4
{
	"path": "var/db/baetyl/app/V1"
}
***********************************/
/**********************************
agent version > 0.1.4
// master will write the ota log to 'var/db/baetyl/ota.log'
// agent will report the content of 'var/db/baetyl/ota.log' to cloud.
// update application
{
	"type": "APP"
	"path": "var/db/baetyl/ota/app/V1"
	"trace": "xxxx-xx-xx-xxxxxxxx"
}
// update master
{
	"type": "MST"
	"path": "var/db/baetyl/ota/mst/0.1.6/baetyl"
	"trace": "xxxx-xx-xx-xxxxxxxx"
}
***********************************/
func (s *Server) updateSystem(_ http.Params, reqBody []byte) ([]byte, error) {
	if reqBody == nil {
		return nil, fmt.Errorf("request body invalid")
	}
	args := make(map[string]string)
	err := json.Unmarshal(reqBody, &args)
	if err != nil {
		return nil, err
	}
	tp, ok := args["type"]
	if !ok {
		tp = "APP"
	}
	target, ok := args["path"]
	if !ok {
		// backward compatibility, agent version < 0.1.4
		target = args["file"]
	}
	trace, _ := args["trace"]
	go s.m.UpdateSystem(trace, tp, target)
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

// deprecated

func (s *Server) inspectSystemV0(_ http.Params, reqBody []byte) ([]byte, error) {
	data, err := s.m.InspectSystem()
	if err != nil {
		return nil, err
	}

	v1 := &baetyl.Inspect{}
	err = json.Unmarshal(data, v1)
	if err != nil {
		return nil, err
	}
	return toInspectSystemV0(v1)
}

func toInspectSystemV0(v1 *baetyl.Inspect) ([]byte, error) {
	v0 := &InspectV0{
		Time:     v1.Time,
		Error:    v1.Error,
		Software: v1.Software,
		Services: v1.Services,
		Volumes:  v1.Volumes,
	}
	v0.Hardware.HostInfo = v1.Hardware.HostInfo
	v0.Hardware.DiskInfo = v1.Hardware.DiskInfo
	v0.Hardware.NetInfo = v1.Hardware.NetInfo
	v0.Hardware.MemInfo = v1.Hardware.MemInfo
	v0.Hardware.CPUInfo = &CPUInfoV0{}
	v0.Hardware.CPUInfo.UsedPercent = v1.Hardware.CPUInfo.UsedPercent
	v0.Hardware.GPUInfo = []GPUInfoV0{}
	for _, v := range v1.Hardware.GPUInfo.GPUs {
		v0.Hardware.GPUInfo = append(v0.Hardware.GPUInfo, GPUInfoV0{
			ID:    v.Index,
			Model: v.Model,
			Mem: utils.MemInfo{
				Total:       v.MemTotal,
				Free:        v.MemFree,
				UsedPercent: v.MemUsedPercent,
			},
		})
	}
	return json.Marshal(v0)
}

// InspectV0 all baetyl information and status inspected
type InspectV0 struct {
	// exception information
	Error string `json:"error,omitempty"`
	// inspect time
	Time time.Time `json:"time,omitempty"`
	// software information
	Software baetyl.Software `json:"software,omitempty"`
	// hardware information
	Hardware HardwareV0 `json:"hardware,omitempty"`
	// service information, including service name, instance running status, etc.
	Services baetyl.Services `json:"services,omitempty"`
	// storage volume information, including name and version
	Volumes baetyl.Volumes `json:"volumes,omitempty"`
}

// HardwareV0 hardware information
type HardwareV0 struct {
	// host information
	HostInfo *utils.HostInfo `json:"host_stats,omitempty"`
	// net information of host
	NetInfo *utils.NetInfo `json:"net_stats,omitempty"`
	// memory usage information of host
	MemInfo *utils.MemInfo `json:"mem_stats,omitempty"`
	// CPU usage information of host
	CPUInfo *CPUInfoV0 `json:"cpu_stats,omitempty"`
	// disk usage information of host
	DiskInfo *utils.DiskInfo `json:"disk_stats,omitempty"`
	// CPU usage information of host
	GPUInfo []GPUInfoV0 `json:"gpu_stats,omitempty"`
}

// CPUInfoV0 CPU information
type CPUInfoV0 struct {
	UsedPercent float64 `json:"used_percent,omitempty"`
}

// GPUInfoV0 GPU information
type GPUInfoV0 struct {
	ID    string        `json:"id,omitempty"`
	Model string        `json:"model,omitempty"`
	Mem   utils.MemInfo `json:"mem_stat,omitempty"`
}
