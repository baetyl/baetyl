package baetyl

import (
	"path/filepath"
)

const licensePath = "license.cert"

type license struct {
	certChain
	serial string
}

func (rt *runtime) loadLicense() error {
	var err error
	rt.license.certChain, err = loadCertChain(filepath.Join(rt.cfg.DataPath, licensePath))
	if err != nil {
		lp := rt.cfg.License
		if !filepath.IsAbs(lp) {
			lp = filepath.Join(rt.cfg.DataPath, lp)
		}
		rt.license.certChain, err = loadCertChain(filepath.Join(rt.cfg.DataPath, licensePath))
	}
	if err != nil {
		return err
	}
	rt.license.serial = rt.license.certChain.cert.Leaf.SerialNumber.String()
	return nil
}
