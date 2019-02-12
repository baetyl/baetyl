package master

import (
	"encoding/json"
	"fmt"

	"github.com/baidu/openedge/logger"
	"github.com/baidu/openedge/protocol/http"
	"github.com/baidu/openedge/sdk-go/openedge"
)

// Server master server to start/stop modules
type Server struct {
	*http.Server
	master *Master
	log    logger.Logger
}

func (m *Master) initServer() error {
	svr, err := http.NewServer(m.inicfg.Server, m.auth)
	if err != nil {
		return err
	}
	s := &Server{
		Server: svr,
		master: m,
		log:    m.log.WithField("master", "server"),
	}
	s.Handle(s.inspect, "GET", "/inspect")
	s.Handle(s.reload, "PUT", "/update")
	m.server = s
	return nil
}

func (s *Server) inspect(_ http.Params, reqBody []byte) ([]byte, error) {
	resBody, err := json.Marshal(s.master.stats())
	if err != nil {
		return nil, err
	}
	return resBody, nil
}

func (s *Server) reload(_ http.Params, reqBody []byte) ([]byte, error) {
	if reqBody == nil {
		return nil, fmt.Errorf("request body invalid")
	}
	d := new(openedge.DatasetInfo)
	err := json.Unmarshal(reqBody, d)
	if err != nil {
		return nil, err
	}
	go s.master.reload(d)
	return nil, nil
}
