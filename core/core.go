package core

import (
	"crypto/x509"
	"fmt"
	"os"
	"strings"

	"github.com/baetyl/baetyl-go/v2/context"
	"github.com/baetyl/baetyl-go/v2/errors"
	"github.com/baetyl/baetyl-go/v2/http"
	"github.com/baetyl/baetyl-go/v2/utils"
	routing "github.com/qiangxue/fasthttp-routing"
	bh "github.com/timshannon/bolthold"
	"github.com/valyala/fasthttp"

	"github.com/baetyl/baetyl/config"
	"github.com/baetyl/baetyl/engine"
	"github.com/baetyl/baetyl/node"
	"github.com/baetyl/baetyl/store"
	"github.com/baetyl/baetyl/sync"
)

type Core struct {
	cfg config.Config
	sto *bh.Store
	sha *node.Node
	eng *engine.Engine
	syn sync.Sync
	svr *http.Server
}

const (
	EnvKeyNodeNamespace = "BAETYL_NODE_NAMESPACE"
)

// NewCore creates a new core
func NewCore(cfg config.Config) (*Core, error) {
	err := extractNodeInfo(cfg)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c := &Core{}
	c.sto, err = store.NewBoltHold(cfg.Store.Path)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.sha, err = node.NewNode(c.sto)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.syn, err = sync.NewSync(cfg, c.sto, c.sha)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.eng, err = engine.NewEngine(cfg, c.sto, c.sha, c.syn)
	if err != nil {
		return nil, errors.Trace(err)
	}
	c.svr = http.NewServer(cfg.Server, c.initRouter())

	c.eng.Start()
	c.syn.Start()
	c.svr.Start()
	return c, nil
}

func (c *Core) Close() {
	if c.svr != nil {
		c.svr.Close()
	}
	if c.eng != nil {
		c.eng.Close()
	}
	if c.sto != nil {
		c.sto.Close()
	}
	if c.syn != nil {
		c.syn.Close()
	}
}

func (c *Core) initRouter() fasthttp.RequestHandler {
	router := routing.New()
	router.Get("/node/stats", c.sha.GetStats)
	router.Get("/services/<service>/log", c.eng.GetServiceLog)
	return router.HandleRequest
}

func extractNodeInfo(cfg config.Config) error {
	tlsConfig, err := utils.NewTLSConfigClient(cfg.Node)
	if err != nil {
		return err
	}
	if len(tlsConfig.Certificates) == 1 && len(tlsConfig.Certificates[0].Certificate) == 1 {
		cert, err := x509.ParseCertificate(tlsConfig.Certificates[0].Certificate[0])
		if err == nil {
			res := strings.SplitN(cert.Subject.CommonName, ".", 2)
			if len(res) != 2 || res[0] == "" || res[1] == "" {
				return fmt.Errorf("failed to parse node name from cert")
			} else {
				os.Setenv(context.KeyNodeName, res[1])
				os.Setenv(EnvKeyNodeNamespace, res[0])
			}
		} else {
			return fmt.Errorf("certificate format error")
		}
	}
	return nil
}
