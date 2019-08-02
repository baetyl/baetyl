package master

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	openedge "github.com/baidu/openedge/sdk/openedge-go"
	"github.com/baidu/openedge/utils"
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
		utils.SetEnv(openedge.EnvMasterHostSocket, sock)
		if c.Mode == "native" {
			utils.SetEnv(openedge.EnvMasterAPIKey, "unix://"+openedge.DefaultSockFile)
		} else {
			utils.SetEnv(openedge.EnvMasterAPIKey, "unix:///"+openedge.DefaultSockFile)
		}
	} else {
		if c.Mode == "native" {
			utils.SetEnv(openedge.EnvMasterAPIKey, addr)
		} else {
			parts := strings.SplitN(url.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
			utils.SetEnv(openedge.EnvMasterAPIKey, addr)
		}
	}
	utils.SetEnv(openedge.EnvMasterAPIVersionKey, "v1")
	utils.SetEnv(openedge.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(openedge.EnvRunningModeKey, c.Mode)
	return nil
}
