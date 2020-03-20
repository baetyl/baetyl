package baetyl

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/tls"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	schema0 "github.com/baetyl/baetyl/schema/v0"
	schema "github.com/baetyl/baetyl/schema/v3"
	"github.com/baetyl/baetyl/utils"

	"gopkg.in/yaml.v2"
)

const updateAppInDesire = "application.yml"
const updateAppFileName = "update.yml"

type updateData struct {
	Type    string       `json:"type,omitempty"`
	Trace   string       `json:"trace,omitempty"`
	Version string       `json:"version,omitempty"`
	Config  updateVolume `json:"config,omitempty"`
}

type updateVolume struct {
	Name string     `json:"name" validate:"regexp=^[a-zA-Z0-9_-]{0\\,63}$"`
	Path string     `json:"path" validate:"nonzero"`
	Meta updateMeta `json:"meta"`
}

type updateMeta struct {
	URL     string `json:"url"`
	MD5     string `json:"md5"`
	Version string `json:"version"`
}

type updateAppConfig struct {
	Version  string                `yaml:"version,omitempty" json:"version,omitempty"`
	Services []schema0.ServiceInfo `yaml:"services,omitempty" json:"services,omitempty" default:"{}"`
	Volumes  []updateVolume        `yaml:"volumes,omitempty" json:"volumes,omitempty" default:"{}"`
}

func (rt *runtime) localUpdate(ctx context.Context) {
	uf := filepath.Join(rt.cfg.DataPath, updateAppFileName)
	data, err := ioutil.ReadFile(uf)
	if err != nil {
		os.Remove(uf)
		return
	}
	uac := &updateAppConfig{}
	err = utils.UnmarshalYAML(data, uac)
	if err != nil {
		os.Remove(uf)
		return
	}
	select {
	case <-ctx.Done():
		return
	default:
	}
	rt.utask.run(func(ctx context.Context) error {
		err := rt.update(ctx, uac)
		if err != nil {
			rt.log.Errorf("local update fail: %s", err.Error())
		}
		return err
	})
}

func (rt *runtime) remoteUpdate(ctx context.Context, ud *updateData, reply func()) error {
	zr, err := rt.peekVolume(ctx, &ud.Config.Meta)
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		if strings.Compare(f.Name, updateAppInDesire) != 0 {
			continue
		}
		r, err := f.Open()
		if err != nil {
			return err
		}
		defer r.Close()
		data, err := ioutil.ReadAll(r)
		if err != nil {
			return err
		}
		var uac updateAppConfig
		err = utils.UnmarshalYAML(data, &uac)
		if err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		rt.utask.run(func(ctx context.Context) error {
			uf := filepath.Join(rt.cfg.DataPath, updateAppFileName)
			data, err := yaml.Marshal(uac)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(uf, data, 0644)
			if err != nil {
				return err
			}
			if reply != nil {
				reply()
			}
			err = rt.update(ctx, &uac)
			if err != nil {
				rt.log.Errorf("remote update fail: %s", err.Error())
			}
			return err
		})
	}
	return nil
}

func (rt *runtime) update(ctx context.Context, uac *updateAppConfig) error {
	log := rt.log.WithField("ota", "update")
	err := rt.downloadUpdate(ctx, uac)
	if err != nil {
		log.Errorf("complete local update fail: %s", err.Error())
		return err
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}
	rt.atask.run(func(ctx context.Context) error {
		ac := schema.ComposeAppConfig{
			AppVersion: uac.Version,
			Services:   map[string]schema.ComposeService{},
			Volumes:    map[string]schema.ComposeVolume{},
		}
		last := ""
		for i, s := range uac.Services {
			depends := make([]string, 0)
			if i > 0 {
				depends = append(depends, last)
			}
			vs := make([]*schema.ServiceVolume, 0)
			for _, m := range s.Mounts {
				vs = append(vs, &schema.ServiceVolume{
					Type:     "bind",
					Source:   m.Name,
					Target:   m.Path,
					ReadOnly: m.ReadOnly,
				})
			}
			ac.Services[s.Name] = schema.ComposeService{
				Image:       s.Image,
				Replica:     s.Replica,
				Volumes:     vs,
				Networks:    s.Networks,
				NetworkMode: s.NetworkMode,
				Ports:       s.Ports,
				Devices:     s.Devices,
				DependsOn:   depends,
				Command: schema.Command{
					Cmd: s.Args,
				},
				Environment: schema.Environment{
					Envs: s.Env,
				},
				Restart:   s.Restart,
				Resources: s.Resources,
				Runtime:   s.Runtime,
			}
			last = s.Name
		}
		data, err := yaml.Marshal(ac)
		if err != nil {
			log.Errorf("marshal appconfig fail: %s", err.Error())
			return err
		}
		configFile := filepath.Join(rt.cfg.DataPath, appConfigFile)
		err = ioutil.WriteFile(configFile, data, 0644)
		if err != nil {
			log.Errorf("write appconfig fail: %s", err.Error())
			return err
		}
		os.Remove(filepath.Join(rt.cfg.DataPath, updateAppFileName))
		select {
		case <-ctx.Done():
			rt.log.Infoln("UPDATE CONTEXT CANCELED")
		default:
		}
		return rt.apply(ctx, &ac)
	})
	return nil
}

func (rt *runtime) downloadUpdate(ctx context.Context, uac *updateAppConfig) error {
	vpath := filepath.Join(rt.cfg.DataPath, injectVolumeDir, uac.Version)
	err := os.MkdirAll(vpath, 0755)
	if err != nil {
		return err
	}
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	for _, vol := range uac.Volumes {
		err = os.MkdirAll(filepath.Join(vpath, vol.Name), 0755)
		if err != nil {
			os.RemoveAll(vpath)
			return err
		}
		if len(vol.Meta.URL) == 0 {
			rt.log.Infof("VOLUME %s EMPTY", vol.Name)
			continue
		}
		vr, err := rt.peekVolume(ctx, &vol.Meta)
		if err != nil {
			os.RemoveAll(vpath)
			return err
		}
		err = rt.extractVolume(vr, filepath.Join(vpath, vol.Name))
		if err != nil {
			os.RemoveAll(vpath)
			return err
		}
		select {
		case <-ctx.Done():
			return nil
		default:
		}
	}
	return nil
}

func (rt *runtime) peekVolume(ctx context.Context, meta *updateMeta) (*zip.Reader, error) {
	sum0, err := base64.StdEncoding.DecodeString(meta.MD5)
	if err != nil {
		return nil, fmt.Errorf("decode ota checksum fail: %s", err.Error())
	}
	rt.log.Infof("begin download volume [url=%s] [checksum=%x]", meta.URL, sum0)
	c := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				RootCAs: rt.license.pool,
			},
		},
		Timeout: rt.cfg.Manage.Desire.Timeout,
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, meta.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header = http.Header{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	pack, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	sum1 := md5.Sum(pack)
	if bytes.Compare(sum0, sum1[:]) != 0 {
		return nil, fmt.Errorf("checksum mismatch [url=%s] [%x] [%x]", meta.URL, sum0, sum1[:])
	}
	return zip.NewReader(bytes.NewReader(pack), int64(len(pack)))
}

func (rt *runtime) extractVolume(r *zip.Reader, base string) error {
	rt.log.Infof("begin extract volume [base=%s]", base)
	for _, f := range r.File {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer rc.Close()

		path := filepath.Join(base, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			f, err := os.OpenFile(
				path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer f.Close()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
