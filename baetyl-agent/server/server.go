package server

import (
	"html/template"
	"net/http"

	"github.com/baetyl/baetyl/logger"
	"icode.baidu.com/baidu/bce-iot/edge-projects/baetyl-activation/config"
)

type activation interface {
	Active(map[string]string) error
}

// Server information about activeService
type Server struct {
	srv   *http.Server
	log   logger.Logger
	cfg   config.Server
	attrs []config.Attribute
	act   activation
}

// NewServer return a instance of activeService
func NewServer(cfg config.Server, attrs []config.Attribute, act activation, log logger.Logger) (*Server, error) {
	server := &Server{
		act:   act,
		cfg:   cfg,
		attrs: attrs,
		log:   log,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", server.handleView)
	mux.HandleFunc("/update", server.handleUpdate)
	srv := &http.Server{}
	srv.Handler = mux
	srv.Addr = cfg.Listen
	server.srv = srv
	return server, nil
}

// Start start activeService
func (s *Server) Start() error {
	s.log.Debugln(s.srv.Addr)
	return s.srv.ListenAndServe()
}

// Close close activeService
func (s *Server) Close() {
	err := s.srv.Close()
	if err != nil {
		s.log.WithError(err).Errorf("failed to close server")
	}
}

func (s *Server) handleView(w http.ResponseWriter, req *http.Request) {
	attrs := map[string][]config.Attribute{"Attributes": s.attrs}
	tpl, err := template.ParseFiles(s.cfg.Pages + "/active.html.template")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, attrs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Server) handleUpdate(w http.ResponseWriter, req *http.Request) {
	s.log.Debugln("handleUpdate")
	if req.Method != http.MethodPost {
		http.Error(w, "post only", http.StatusMethodNotAllowed)
		return
	}
	err := req.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	attributes := make(map[string]string)
	for _, attr := range s.attrs {
		val := req.Form.Get(attr.Name)
		if val == "" {
			attributes[attr.Name] = attr.Value
		} else {
			attributes[attr.Name] = val
		}
	}
	s.log.Debugln("attributes fetch: ", attributes)
	var tpl *template.Template
	page := "/success.html.template"
	err = s.act.Active(attributes)
	if err != nil {
		page = "/failed.html.template"
		s.log.Errorf("Failed to active: %s", err.Error())
	}
	tpl, err = template.ParseFiles(s.cfg.Pages + page)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tpl.Execute(w, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
