package master

import (
	"fmt"
	"io/ioutil"
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
	grpcAddr := c.API.Address
	grpcUrl, err := utils.ParseURL(grpcAddr)
	if err != nil {
		return fmt.Errorf("failed to parse address of API server: %s", err.Error())
	}

	if runtime.GOOS != "linux" && (url.Scheme == "unix" || grpcUrl.Scheme == "unix") {
		return fmt.Errorf("unix domain socket only support on linux, please to use tcp socket")
	}
	if (url.Scheme != "unix" && url.Scheme != "tcp") ||
		(grpcUrl.Scheme != "unix" && grpcUrl.Scheme != "tcp") {
		return fmt.Errorf("only support unix domian socket or tcp socket")
	}

	// address in container
	if url.Scheme == "unix" {
		// http
		sock, err := filepath.Abs(url.Host)
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Dir(sock), 0755)
		if err != nil {
			return err
		}
		utils.SetEnv(baetyl.EnvKeyMasterAPISocket, sock)
		unixPrefix := "unix:"
		if c.Mode == "native" {
			unixPrefix += "/"
		} else {
			unixPrefix += "//"
		}
		utils.SetEnv(baetyl.EnvKeyMasterAPIAddress, unixPrefix+sock)
		// TODO: remove, backward compatibility
		utils.SetEnv(baetyl.EnvMasterAPIKey, unixPrefix+sock)

		// grpc
		grpcSock, err := filepath.Abs(grpcUrl.Host)
		if err != nil {
			return err
		}
		err = os.MkdirAll(filepath.Dir(grpcSock), 0755)
		if err != nil {
			return err
		}
		utils.SetEnv(baetyl.EnvKeyMasterGRPCAPISocket, grpcSock)
		utils.SetEnv(baetyl.EnvKeyMasterGRPCAPIAddress, unixPrefix+grpcSock)
	} else {
		if c.Mode != "native" {
			parts := strings.SplitN(url.Host, ":", 2)
			addr = fmt.Sprintf("tcp://host.docker.internal:%s", parts[1])
			parts = strings.SplitN(grpcUrl.Host, ":", 2)
			grpcUrl.Host = fmt.Sprintf("host.docker.internal:%s", parts[1])
		}
		utils.SetEnv(baetyl.EnvKeyMasterAPIAddress, addr)
		utils.SetEnv(baetyl.EnvKeyMasterGRPCAPIAddress, grpcUrl.Host)
		// TODO: remove, backward compatibility
		utils.SetEnv(baetyl.EnvMasterAPIKey, addr)
	}

	if c.SNFile != "" {
		snByte, err := ioutil.ReadFile(c.SNFile)
		if err != nil {
			fmt.Printf("failed to load SN file: %s", err.Error())
		} else {
			sn := strings.TrimSpace(string(snByte))
			utils.SetEnv(baetyl.EnvKeyHostSN, sn)
		}
	}

	utils.SetEnv(baetyl.EnvKeyMasterAPIVersion, "v1")
	utils.SetEnv(baetyl.EnvKeyHostOS, runtime.GOOS)
	utils.SetEnv(baetyl.EnvKeyServiceMode, c.Mode)
	// TODO: remove, backward compatibility
	utils.SetEnv(baetyl.EnvMasterAPIVersionKey, "v1")
	utils.SetEnv(baetyl.EnvHostOSKey, runtime.GOOS)
	utils.SetEnv(baetyl.EnvRunningModeKey, c.Mode)

	hi := utils.GetHostInfo()
	if hi.HostID != "" {
		utils.SetEnv(baetyl.EnvKeyHostID, hi.HostID)
		// TODO: remove, backward compatibility
		utils.SetEnv(baetyl.EnvHostID, hi.HostID)
	}
	return nil
}
