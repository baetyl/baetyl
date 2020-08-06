package initz

import (
	"context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
	"github.com/baetyl/baetyl/config"
	"html/template"
	"net/http"
)

func (active *Activate) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", active.handleView)
	mux.HandleFunc("/update", active.handleUpdate)
	srv := &http.Server{}
	srv.Handler = mux
	srv.Addr = active.cfg.Init.ActivateConfig.Server.Listen
	active.srv = srv
	return errors.Trace(active.srv.ListenAndServe())
}

func (active *Activate) closeServer() {
	err := active.srv.Shutdown(context.Background())
	if err != nil {
		active.log.Error("active", log.Any("server err", err))
	}
}

func (active *Activate) handleView(w http.ResponseWriter, req *http.Request) {
	attrs := map[string][]config.Attribute{"Attributes": active.cfg.Init.ActivateConfig.Attributes}
	tpl, err := template.ParseFiles(active.cfg.Init.ActivateConfig.Server.Pages + "/active.html.template")
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

func (active *Activate) handleUpdate(w http.ResponseWriter, req *http.Request) {
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
	for _, attr := range active.cfg.Init.ActivateConfig.Attributes {
		val := req.Form.Get(attr.Name)
		if val == "" {
			attributes[attr.Name] = attr.Value
		} else {
			attributes[attr.Name] = val
		}
	}
	active.log.Info("active", log.Any("server attrs", attributes))
	active.attrs = attributes

	var tpl *template.Template
	page := "/success.html.template"
	active.activate()
	if !utils.FileExists(active.cfg.Sync.Cloud.HTTP.Cert) {
		page = "/failed.html.template"
	}
	tpl, err = template.ParseFiles(active.cfg.Init.ActivateConfig.Server.Pages + page)
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
