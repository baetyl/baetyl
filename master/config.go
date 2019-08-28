package master

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
	"github.com/baetyl/baetyl/utils"
)

// Validate validates config
// TODO: it is not good idea to set envs here
func (c *Config) Validate() error {
	addr := c.Server.Address
	url, err := utils.ParseURL(addr)
	if err != nil {
		return fmt.Errorf("failed to parse address of server: %s", err.Error())
	}

	if runtime.GOOS != "linux" && url.Scheme == "unix" {
		return fmt.Errorf("unix domain socket only support on linux, please to use tcp socket")
	}
	if url.Scheme != "unix" && url.Scheme != "tcp" {
		return fmt.Errorf("only support unix domian socket or tcp socket")
	}

	// address in container
	if url.Scheme == "unix" {
		sock, err := filepath.Abs(url.Host)
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Dir(sock), 0755)
		if err != nil {
			return err
		}
		utils.SetEnv(baetyl.EnvMasterHostSocket, sock)
		if c.Mode == "native" {
			utils.SetEnv(baetyl.EnvMasterAPIKey, "unix://"+baetyl.DefaultSockFile)
		} else {
			utils.SetEnv(baetyl.EnvMasterAPIKey, "unix:///"+baetyl.DefaultSockFile)
		}
	} else {
		if c.Mode == "native" {
			utils.SetEnv(baetyl.EnvMasterAPIKey, addr)
		} else {
			parts := strings.SplitN(url.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
			utils.SetEnv(baetyl.EnvMasterAPIKey, addr)
		}
	}
	utils.SetEnv(baetyl.EnvMasterAPIVersionKey, "v1")
	utils.SetEnv(baetyl.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(baetyl.EnvRunningModeKey, c.Mode)

	hi := utils.GetHostInfo()
	if hi.HostID != "" {
		utils.SetEnv(baetyl.EnvHostID, hi.HostID)
	}
	return nil
}
