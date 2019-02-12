package api

import (
	"encoding/json"
	"fmt"

	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/sdk-go/openedge"
)

// Master master inerface
type Master interface {
	Auth(u, p string) bool
	Inspect() *openedge.Inspect
	Update(*openedge.DatasetInfo) error
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
	s.s.Handle(s.inspect, "GET", "/inspect")
	s.s.Handle(s.update, "PUT", "/update")
	return s, s.s.Start()
}

// Close closes api server
func (s *Server) Close() error {
	return s.s.Close()
}

func (s *Server) inspect(_ http.Params, reqBody []byte) ([]byte, error) {
	resBody, err := json.Marshal(s.m.Inspect())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (s *Server) update(_ http.Params, reqBody []byte) ([]byte, error) {
	if reqBody == nil {
		return nil, fmt.Errorf("request body invalid")
	}
	d := new(openedge.DatasetInfo)
	err := json.Unmarshal(reqBody, d)
	if err != nil {
		return nil, err
	}
	go s.m.Update(d)
	return nil, nil
}
