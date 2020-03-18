package baetyl

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/baetyl/baetyl/utils"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func (rt *runtime) runServer(ctx context.Context) error {
	if rt.cfg.APIConfig.Network == "unix" {
		os.Remove(rt.cfg.APIConfig.Address)
	}
	rt.srv = grpc.NewServer(grpc.Creds(credentials.NewTLS(&tls.Config{
		Certificates: []tls.Certificate{rt.cert.cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    rt.cert.pool,
	})))
	// FIXME RegisterKVServiceServer(rt.svr, NewKVService(rt.db))
	lis, err := net.Listen(rt.cfg.APIConfig.Network, rt.cfg.APIConfig.Address)
	if err != nil {
		return err
	}
	go rt.srv.Serve(lis)
	rt.log.Infoln("apiserver started")

	<-ctx.Done()

	rt.srv.GracefulStop()
	rt.log.Infoln("apiserver stopped")
	return ctx.Err()
}

type legacyServer struct {
	http.Server
	router *mux.Router
}

func (rt *runtime) runLegacyServer(ctx context.Context) error {
	rt.lsrv.WriteTimeout = rt.cfg.LegacyAPIConfig.Timeout
	rt.lsrv.ReadTimeout = rt.cfg.LegacyAPIConfig.Timeout
	rt.lsrv.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{rt.cert.cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    rt.cert.pool,
	}
	rt.lsrv.router = mux.NewRouter()
	rt.lsrv.Handler = rt.lsrv.router

	// v0, deprecated
	rt.handle(rt.onInspectSystemV0, "GET", "/system/inspect")
	rt.handle(rt.onGetAvailablePort, "GET", "/ports/available")

	// v1
	rt.handle(rt.onInspectSystem, "GET", "/v1/system/inspect")
	rt.handle(rt.onGetAvailablePort, "GET", "/v1/ports/available")

	if rt.cfg.LegacyAPIConfig.Network == "unix" {
		os.Remove(rt.cfg.LegacyAPIConfig.Address)
	}
	lis, err := net.Listen(rt.cfg.LegacyAPIConfig.Network, rt.cfg.LegacyAPIConfig.Address)
	if err != nil {
		return err
	}
	if rt.cfg.LegacyAPIConfig.Network == "tcp" {
		lis = tcpKeepAliveListener{lis.(*net.TCPListener)}
	}
	go rt.lsrv.Serve(lis)
	rt.log.Infoln("legacy apiserver started")

	<-ctx.Done()

	ctx2, cancel := context.WithTimeout(context.Background(), rt.lsrv.IdleTimeout)
	defer rt.log.Infoln("legacy apiserver stopped")
	defer cancel()
	rt.lsrv.Shutdown(ctx2)
	return ctx.Err()
}

func (rt *runtime) handle(handle func(map[string]string, []byte) ([]byte, error), method, path string, params ...string) {
	rt.lsrv.router.HandleFunc(path, func(res http.ResponseWriter, req *http.Request) {
		rt.log.Infof("[%s] %s", req.Method, req.URL.String())
		var err error
		var reqBody []byte
		if req.Body != nil {
			defer req.Body.Close()
			reqBody, err = ioutil.ReadAll(req.Body)
			if err != nil {
				http.Error(res, err.Error(), 400)
				rt.log.Errorf("[%s] %s %s", req.Method, req.URL.String(), err.Error())
				return
			}
		}
		resBody, err := handle(mux.Vars(req), reqBody)
		if err != nil {
			http.Error(res, err.Error(), 400)
			rt.log.Errorf("[%s] %s %s", req.Method, req.URL.String(), err.Error())
			return
		}
		if resBody != nil {
			res.Write(resBody)
		}
	}).Methods(method).Queries(params...)
}

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (ln tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := ln.AcceptTCP()
	if err != nil {
		return nil, err
	}
	tc.SetKeepAlive(true)
	tc.SetKeepAlivePeriod(3 * time.Minute)
	return tc, nil
}

func (rt *runtime) onInspectSystem(_ map[string]string, reqBody []byte) ([]byte, error) {
	return json.Marshal(rt.inspect())
}

func (rt *runtime) onGetAvailablePort(_ map[string]string, reqBody []byte) ([]byte, error) {
	port, err := utils.GetAvailablePort("127.0.0.1")
	if err != nil {
		return nil, err
	}
	res := make(map[string]string)
	res["port"] = strconv.Itoa(port)
	return json.Marshal(res)
}

func (rt *runtime) onInspectSystemV0(_ map[string]string, reqBody []byte) ([]byte, error) {
	return json.Marshal(rt.inspect().ToV0())
}
