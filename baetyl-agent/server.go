package main

import (
	"github.com/baetyl/baetyl/baetyl-agent/config"
	"github.com/baetyl/baetyl/logger"
	"html/template"
	"net/http"
)

type Server struct {
	server *http.Server
	config config.Server
	log    logger.Logger
}

func (a *agent) NewServer(config config.Server, log logger.Logger) error {
	s := &Server{
		config: config,
		log:    log,
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", a.handleView)
	mux.HandleFunc("/update", a.handleUpdate)
	srv := &http.Server{}
	srv.Handler = mux
	srv.Addr = config.Listen
	s.server = srv
	a.srv = s
	return nil
}

// Start start activeService
func (a *agent) StartServer() error {
	a.ctx.Log().Debugln(a.srv.server.Addr)
	return a.srv.server.ListenAndServe()
}

// Close close activeService
func (a *agent) CloseServer() {
	err := a.srv.server.Close()
	if err != nil {
		a.srv.log.WithError(err).Errorf("failed to close server")
	}
}

func (a *agent) handleView(w http.ResponseWriter, req *http.Request) {
	attrs := map[string][]config.Attribute{"Attributes": a.cfg.Attributes}
	tpl, err := template.ParseFiles(a.cfg.Server.Pages + "/active.html.template")
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

func (a *agent) handleUpdate(w http.ResponseWriter, req *http.Request) {
	a.srv.log.Debugln("handleUpdate")
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
	for _, attr := range a.cfg.Attributes {
		val := req.Form.Get(attr.Name)
		if val == "" {
			attributes[attr.Name] = attr.Value
		} else {
			attributes[attr.Name] = val
		}
	}
	a.srv.log.Debugln("attributes fetch: ", attributes)
	a.attrs = attributes

	var tpl *template.Template
	page := "/success.html.template"
	a.report()
	if a.node == nil {
		page = "/failed.html.template"
		a.srv.log.Errorf("Failed to active: %s", err.Error())
	} else {
		err := a.tomb.Go(a.reporting, a.processing)
		if err != nil {
			a.srv.log.WithError(err)
		}
	}
	tpl, err = template.ParseFiles(a.cfg.Server.Pages + page)
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
