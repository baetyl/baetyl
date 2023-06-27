package initz

import (
	"context"
	"html/template"
	"net/http"
	"os"

	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/log"
	"github.com/baetyl/baetyl-go/v2/utils"
)

const (
	KeyBaetylSyncAddr = "BAETYL_SYNC_ADDR"
)

func (active *Activate) startServer() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", active.handleView)
	mux.HandleFunc("/update", active.handleUpdate)
	mux.HandleFunc("/active", active.handleActive)
	srv := &http.Server{}
	srv.Handler = mux
	srv.Addr = active.cfg.Init.Active.Collector.Server.Listen
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
	attrs := map[string]interface{}{
		"Attributes": active.cfg.Init.Active.Collector.Attributes,
		"Nodeinfo":   active.cfg.Init.Active.Collector.NodeInfo,
		"Serial":     active.cfg.Init.Active.Collector.Serial,
	}
	tpl, err := template.ParseFiles(active.cfg.Init.Active.Collector.Server.Pages + "/active.html.template")
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
	for _, attr := range active.cfg.Init.Active.Collector.Attributes {
		val := req.Form.Get(attr.Name)
		if val == "" {
			attributes[attr.Name] = attr.Value
		} else {
			attributes[attr.Name] = val
		}
	}
	for _, ni := range active.cfg.Init.Active.Collector.NodeInfo {
		val := req.Form.Get(ni.Name)
		attributes[ni.Name] = val
	}
	for _, si := range active.cfg.Init.Active.Collector.Serial {
		val := req.Form.Get(si.Name)
		attributes[si.Name] = val
	}
	active.log.Info("active", log.Any("server attrs", attributes))
	active.attrs = attributes

	if batchName, ok := attributes["batch"]; ok {
		active.batch.name = batchName
	}
	if ns, ok := attributes["namespace"]; ok {
		active.batch.namespace = ns
	}
	if initAddr, ok := attributes["initAddr"]; ok {
		active.cfg.Init.Active.Address = initAddr

	}
	if syncAddr, ok := attributes["syncAddr"]; ok {
		os.Setenv(KeyBaetylSyncAddr, syncAddr)
	}

	var tpl *template.Template
	page := "/success.html.template"
	active.activate()
	if !utils.FileExists(active.cfg.Node.Cert) {
		page = "/failed.html.template"
	}
	tpl, err = template.ParseFiles(active.cfg.Init.Active.Collector.Server.Pages + page)
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

func (active *Activate) handleActive(w http.ResponseWriter, req *http.Request) {
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
	for _, attr := range active.cfg.Init.Active.Collector.Attributes {
		val := req.Form.Get(attr.Name)
		if val == "" {
			attributes[attr.Name] = attr.Value
		} else {
			attributes[attr.Name] = val
		}
	}
	for _, ni := range active.cfg.Init.Active.Collector.NodeInfo {
		val := req.Form.Get(ni.Name)
		attributes[ni.Name] = val
	}
	for _, si := range active.cfg.Init.Active.Collector.Serial {
		val := req.Form.Get(si.Name)
		attributes[si.Name] = val
	}
	active.log.Info("active", log.Any("server attrs", attributes))
	active.attrs = attributes

	if batchName, ok := attributes["batch"]; ok {
		active.batch.name = batchName
	}
	if ns, ok := attributes["namespace"]; ok {
		active.batch.namespace = ns
	}
	if initAddr, ok := attributes["initAddr"]; ok {
		active.cfg.Init.Active.Address = initAddr

	}
	if syncAddr, ok := attributes["syncAddr"]; ok {
		os.Setenv(KeyBaetylSyncAddr, syncAddr)
	}

	err = active.activate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write([]byte("active success"))
}
