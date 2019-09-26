# Customize Module

Read [Build Baetyl From Source](../install/Build-from-Source.md) before developing custom modules to understand Baetyl's build environment requirements.

Custom modules do not limit the development language. Understand these conventions below to integrate custom modules better and faster.

The custom module does not limit the development language. As long as it is a runnable program, you can even use the image already on hub.docker.com, such as `eclipse-mosquitto`. But understanding the conventions described below will help you develop custom modules better and faster.

## Directory Convention

At present, the native process mode, like the docker container mode, opens up a separate workspace for each service. Although it does not achieve the effect of isolation, it can guarantee the consistency of the user experience. The process mode creates a separate directory for each service in the `var/run/baetyl/services` directory, using service name. When the server starts, it specifies the directory as the working directory, and the service-bound storage volumes will be mapped (soft link) to the working directory. Here we keep the definition of the docker container mode, the workspace under the directory is also called the container, then the directory in the container has the following recommended usage:

- Default working directory in the container: `/`
- Default configuration file in the container: `/etc/baetyl/service.yml`
- Default persistence path in the container: `/var/db/baetyl`
- Default log directory in the container: `/var/log/baetyl`

**NOTE**: If the data needs to be persisted on the device (host), such as database and log, the directory in the container must be mapped to the host directory through the storage volume, otherwise the data will be lost after the service is stopped.

## Start/Stop Convention

There is no excessive requirement for the module to be started. But it is recommended to load the YMAL format configuration from the default file, then run the module's business logic, and finally listen to the `SIGTERM` signal to gracefully exit. A simple `Golang` module implementation can refer to the MQTT remote communication module (`baetyl-remote-mqtt`).

## SDK

If the module is developed using `Golang`, you can use the SDK provided by Baetyl, located in the sdk directory of the project, and the functional interfaces are provided by `Context`. At present, the SDK capabilities provided are still not enough, and the follow-up will be gradually strengthened.

The list of `Context` interfaces are as follows:

```golang
// returns the system configuration of the service, such as hub and logger
Config() *ServiceConfig
// loads the custom configuration of the service
LoadConfig(interface{}) error
// creates a Client that connects to the Hub through system configuration,
// you can specify the Client ID and the topic information of the subscription.
NewHubClient(string, []mqtt.TopicInfo) (*mqtt.Dispatcher, error)
// returns logger interface
Log() logger.Logger
// check running mode
IsNative() bool
// waiting to exit, receiving SIGTERM and SIGINT signals
Wait()
// returns wait channel
WaitChan() <-chan os.Signal

// Master RESTful API

// updates system and
UpdateSystem(string, bool) error
// inspects system stats
InspectSystem() (*Inspect, error)
// gets an available port of the host
GetAvailablePort() (string, error)
// reports the stats of the instance of the service
ReportInstance(stats map[string]interface{}) error
// starts an instance of the service
StartInstance(serviceName, instanceName string, dynamicConfig map[string]string) error
// stop the instance of the service
StopInstance(serviceName, instanceName string) error
```

The following uses the simple timer module implementation as an example to introduce the usage of the SDK.

```golang
package main

import (
	"encoding/json"
	"time"

	"github.com/baetyl/baetyl/protocol/mqtt"
	baetyl "github.com/baetyl/baetyl/sdk/baetyl-go"
)

// custom configuration of the timer module
type config struct {
	Timer struct {
		Interval time.Duration `yaml:"interval" json:"interval" default:"1m"`
	} `yaml:"timer" json:"timer"`
	Publish mqtt.TopicInfo `yaml:"publish" json:"publish" default:"{\"topic\":\"timer\"}"`
}

func main() {
	// Running module in baetyl context
	baetyl.Run(func(ctx baetyl.Context) error {
		var cfg config
		// load custom config
		err := ctx.LoadConfig(&cfg)
		if err != nil {
			return err
		}
		// create a hub client
		cli, err := ctx.NewHubClient("", nil)
		if err != nil {
			return err
		}
		// start client to keep connection with hub
		cli.Start(nil)
		// create a timer
		ticker := time.NewTicker(cfg.Timer.Interval)
		defer ticker.Stop()
		for {
			select {
			case t := <-ticker.C:
				msg := map[string]int64{"time": t.Unix()}
				pld, _ := json.Marshal(msg)
				// send a message to hub triggered by timer
				err := cli.Publish(cfg.Publish, pld)
				if err != nil {
					// log error message
					ctx.Log().Errorf(err.Error())
				}
			case <-ctx.WaitChan():
				// wait until service is stopped
				return nil
			}
		}
	})
}
```
