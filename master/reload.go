package master

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/baidu/openedge/sdk-go/openedge"
	"github.com/baidu/openedge/utils"
	"github.com/baidubce/bce-sdk-go/http"
	"github.com/baidubce/bce-sdk-go/util"
	"github.com/docker/distribution/uuid"
	"github.com/mholt/archiver"
)

var datasetsDir = path.Join("var", "db", "openedge")
var configFile = path.Join(datasetsDir, "config.yml")
var downloadDir = path.Join("var", "run", "openedge", "download")

func (m *Master) reload(d *openedge.DatasetInfo) error {
	//  download dataset as config for master
	dir, err := m.download(d)
	if err != nil {
		return fmt.Errorf("failed to download dataset (%s:%s): %s", d.Name, d.Version, err.Error())
	}
	// parse config of next version
	nxt, err := m.parse(dir)
	if err != nil {
		return err
	}
	// diff config of next version with config of current version used by mater
	res, same := nxt.diff(&m.curcfg)
	if same {
		m.log.Infof("config not changed, return directly")
		return nil
	}
	// download all new datasets
	for _, d := range res.addDatasets {
		dir, err = m.download(&d)
		if err != nil {
			return fmt.Errorf("failed to download dataset (%s:%s): %s", d.Name, d.Version, err.Error())
		}
	}
	// stop removed services
	m.stopServices(res.stopServices)
	// start added services
	err = m.startServices(res.startServices)
	if err != nil {
		m.log.Infof("failed to start added services, to rollback")
		defer
		// rollback
		// stop added services
		m.stopServices(res.startServices)
		// start removed services
		err1 := m.startServices(res.stopServices)
		if err1 != nil {
			return fmt.Errorf("%s; failed to rollback: %s", err.Error(), err1.Error())
		}
		return err
	}
	return nil
}

func (m *Master) download(d *openedge.DatasetInfo) (string, error) {
	datasetDir := path.Join(datasetsDir, d.Name, d.Version)
	datasetMD5File := path.Join(datasetDir, ".md5")
	datasetZipFile := path.Join(downloadDir, d.Name, d.Version, uuid.Generate().String())
	if utils.FileExists(datasetMD5File) {
		datasetMD5, err := ioutil.ReadFile(datasetMD5File)
		if err != nil {
			return "", err
		}
		if string(datasetMD5) != d.MD5 {
			return "", fmt.Errorf("dateset (%s) exists which MD5 is not expected", d.Name)
		}
		return datasetDir, nil
	}

	req := new(http.Request)
	req.SetUri(d.URL)
	res, err := http.Execute(req)
	if err != nil {
		return "", err
	}
	body := res.Body()
	defer body.Close()

	data, err := ioutil.ReadAll(body)
	if err != nil {
		return "", err
	}

	stream := bytes.NewBuffer(data)
	datasetMD5, err := util.CalculateContentMD5(stream, int64(len(data)))
	if err != nil {
		return "", err
	}
	if datasetMD5 != d.MD5 {
		if datasetMD5 != d.MD5 {
			return "", fmt.Errorf("dateset (%s) donwloaded which MD5 is not expected", d.Name)
		}
	}

	err = ioutil.WriteFile(datasetZipFile, data, os.ModePerm)
	if err != nil {
		os.RemoveAll(datasetDir)
		return "", err
	}
	defer os.RemoveAll(datasetZipFile)

	err = archiver.Zip.Open(datasetZipFile, datasetDir)
	if err != nil {
		os.RemoveAll(datasetDir)
		return "", err
	}
	err = ioutil.WriteFile(datasetMD5File, []byte(d.MD5), os.ModePerm)
	if err != nil {
		os.RemoveAll(datasetDir)
		return "", err
	}
	return datasetDir, nil
}

func (m *Master) parse(dir string) (*DynamicConfig, error) {
	cfgFile := path.Join(dir, "config.yml")
	cfg := new(DynamicConfig)
	err := utils.LoadYAML(cfgFile, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// func (m *Master) check(cfg *engine.ServiceInfo) {
// 	if runtime.GOOS == "linux" && cfg.Resources.CPU.Cpus > 0 {
// 		sysInfo := sysinfo.New(true)
// 		if !sysInfo.CPUCfsPeriod || !sysInfo.CPUCfsQuota {
// 			m.log.Warnf("configuration 'resources.cpu.cpus' of service (%s) is ignored \
// 			because host kernel does not support CPU cfs period/quota or the cgroup is not mounted.",
// 			cfg.Name)
// 			cfg.Resources.CPU.Cpus = 0
// 		}
// 	}
// }
