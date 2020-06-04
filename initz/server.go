package initz

import (
	"html/template"
	"net/http"

	"github.com/baetyl/baetyl-core/config"
	"github.com/baetyl/baetyl-go/errors"
	"github.com/baetyl/baetyl-go/log"
	"github.com/baetyl/baetyl-go/utils"
)

func (init *Initialize) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", init.handleView)
	mux.HandleFunc("/update", init.handleUpdate)
	srv := &http.Server{}
	srv.Handler = mux
	srv.Addr = init.cfg.Init.ActivateConfig.Server.Listen
	init.srv = srv
	return errors.Trace(init.srv.ListenAndServe())
}

func (init *Initialize) closeServer() {
	err := init.srv.Close()
	if err != nil {
		init.log.Error("init", log.Any("server err", err))
	}
}

func (init *Initialize) handleView(w http.ResponseWriter, req *http.Request) {
	attrs := map[string][]config.Attribute{"Attributes": init.cfg.Init.ActivateConfig.Attributes}
	tpl, err := template.ParseFiles(init.cfg.Init.ActivateConfig.Server.Pages + "/active.html.template")
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

func (init *Initialize) handleUpdate(w http.ResponseWriter, req *http.Request) {
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
	for _, attr := range init.cfg.Init.ActivateConfig.Attributes {
		val := req.Form.Get(attr.Name)
		if val == "" {
			attributes[attr.Name] = attr.Value
		} else {
			attributes[attr.Name] = val
		}
	}
	init.log.Info("init", log.Any("server attrs", attributes))
	init.attrs = attributes

	var tpl *template.Template
	page := "/success.html.template"
	init.activate()
	if !utils.FileExists(init.cfg.Sync.Cloud.HTTP.Cert) {
		page = "/failed.html.template"
	}
	tpl, err = template.ParseFiles(init.cfg.Init.ActivateConfig.Server.Pages + page)
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
