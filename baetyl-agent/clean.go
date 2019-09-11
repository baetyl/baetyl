package main

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/baetyl/baetyl/logger"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

type cleaner struct {
	prefix   string
	target   string
	lversion string              // ast version
	lvolumes []baetyl.VolumeInfo // last volumes
	log      logger.Logger
}

func newCleaner(prefix, target string, log logger.Logger) *cleaner {
	return &cleaner{
		prefix: prefix,
		target: target,
		log:    log,
	}
}

func (c *cleaner) reset() {
	c.lversion = ""
	c.lvolumes = nil
}

func (c *cleaner) set(version string, volumes []baetyl.VolumeInfo) {
	c.lversion = version
	c.lvolumes = volumes
}

func (c *cleaner) do(version string) {
	if c.lvolumes == nil || version == "" || c.lversion != version {
		c.log.Debugf("version (%s) is ignored", version)
		return
	}

	c.log.Infof("start to clean '%s'", c.target)
	defer utils.Trace("end to clean,", c.log.Infof)()

	// list folders to remove
	remove, err := list(c.prefix, c.target, c.lvolumes)
	if err != nil {
		c.log.WithError(err).Warnf("failed to list old volumes")
		return
	}
	for _, v := range remove {
		if os.RemoveAll(v) == nil {
			c.log.Infof("old volume is removed: %s", v)
		} else {
			c.log.WithError(err).Warnf("failed to remove old volumes")
		}
	}
}

func list(prefix, target string, volumes []baetyl.VolumeInfo) ([]string, error) {
	keep := map[string]bool{}
	for _, v := range volumes {
		// remove prefix from path
		p, err := filepath.Rel(prefix, v.Path)
		if err != nil {
			continue
		}
		ps := strings.Split(p, string(filepath.Separator))
		if len(ps) == 0 {
			// ignore the case that path equals prefix
			continue
		}
		if ps[0] == ".." {
			// ignore the case that path out of prefix
			continue
		}
		keep[ps[0]] = len(ps) > 1
	}
	infos, err := ioutil.ReadDir(target)
	if err != nil {
		return nil, err
	}
	remove := []string{}
	for _, info := range infos {
		// skip the files and only clean folders,
		if !info.IsDir() {
			continue
		}
		next, ok := keep[info.Name()]
		if !ok {
			remove = append(remove, filepath.Join(target, info.Name()))
		} else if next {
			nextremove, err := list(path.Join(prefix, info.Name()), path.Join(target, info.Name()), volumes)
			if err != nil {
				return nil, err
			}
			remove = append(remove, nextremove...)
		}
	}
	return remove, err
}
