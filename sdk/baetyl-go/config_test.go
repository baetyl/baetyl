package baetyl

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestLoadComposeAppConfigCompatible(t *testing.T) {
	cfg := ComposeAppConfig{
		Version:    "3",
		AppVersion: "v2",
		Services: map[string]ComposeService{
			"test-hub": ComposeService{
				Image: "hub.test.com/baetyl/baetyl-hub",
				Networks: NetworksInfo{
					ServiceNetworks: map[string]ServiceNetwork{
						"test-network": ServiceNetwork{
							Aliases:     []string{"test-net"},
							Ipv4Address: "192.168.0.3",
						},
					},
				},
				Replica:   1,
				Ports:     []string{"1883:1883"},
				Devices:   []string{},
				DependsOn: []string{},
				Restart: RestartPolicyInfo{
					Policy: "always",
					Backoff: BackoffInfo{
						Min:    time.Second,
						Max:    time.Minute * 5,
						Factor: 2,
					},
				},
				Volumes: []ServiceVolume{
					ServiceVolume{
						Source:   "var/db/baetyl/test-hub-conf",
						Target:   "/etc/baetyl",
						ReadOnly: false,
					},
				},
				Command: Command{
					Cmd: []string{"-c", "conf/conf.yml"},
				},
				Entrypoint: Entrypoint{
					Entry: []string{"test"},
				},
				Environment: Environment{
					Envs: map[string]string{
						"version": "v1",
					},
				},
			},
			"test-timer": ComposeService{
				Image:     "hub.test.com/baetyl/baetyl-timer",
				DependsOn: []string{"test-hub"},
				Replica:   1,
				Ports:     []string{},
				Devices:   []string{},
				Environment: Environment{
					Envs: map[string]string{
						"version": "v2",
					},
				},
				Networks: NetworksInfo{
					ServiceNetworks: map[string]ServiceNetwork{
						"test-network": ServiceNetwork{},
					},
				},
				Volumes: []ServiceVolume{
					ServiceVolume{
						Source:   "var/db/baetyl/test-timer-conf",
						Target:   "/etc/baetyl",
						ReadOnly: true,
					},
				},
				Restart: RestartPolicyInfo{
					Policy: "always",
					Backoff: BackoffInfo{
						Min:    time.Second,
						Max:    time.Minute * 5,
						Factor: 2,
					},
				},
				Command: Command{
					Cmd: []string{"/bin/sh"},
				},
				Entrypoint:Entrypoint{
					Entry: []string{"test"},
				},
			},
		},
		Volumes: map[string]ComposeVolume{},
		Networks: map[string]ComposeNetwork{
			"test-network": ComposeNetwork{
				Driver:     "bridge",
				DriverOpts: map[string]string{},
				Labels:     map[string]string{},
			},
		},
	}

	composeConfString := `
version: '3'
app_version: v2
services:
  test-hub:
    image: hub.test.com/baetyl/baetyl-hub
    networks:
      test-network:
        aliases:
          - test-net
        ipv4_address: 192.168.0.3
    replica: 1
    ports:
      - 1883:1883
    volumes:
      - var/db/baetyl/test-hub-conf:/etc/baetyl
    command:
      - '-c'
      - conf/conf.yml
    entrypoint: test
    environment:
      - version=v1
  test-timer:
    image: hub.test.com/baetyl/baetyl-timer
    depends_on:
      - test-hub
    replica: 1
    networks:
      - test-network
    volumes:
      - source: var/db/baetyl/test-timer-conf
        target: /etc/baetyl
        read_only: true
    environment:
      version: v2
    command: '/bin/sh'
    entrypoint: test

networks:
  test-network:
`
	confString := `
version: v2
services:
  - name: test-hub
    image: hub.test.com/baetyl/baetyl-hub
    networks:
      test-network:
        aliases:
          - test-net
        ipv4_address: 192.168.0.3
    replica: 1
    ports:
      - 1883:1883
    mounts:
      - name: test-hub-conf
        path: etc/baetyl
    args:
      - '-c'
      - conf/conf.yml
    entrypoint:
      - 'test'
    env:
      version: v1
  - name: test-timer
    image: hub.test.com/baetyl/baetyl-timer
    replica: 1
    networks:
      - test-network
    mounts:
      - name: test-timer-conf
        path: etc/baetyl
        readonly: true
    env:
      version: v2
    args:
      - '/bin/sh'
    entrypoint:
      - 'test'

networks:
  test-network:

volumes:
  - name: test-hub-conf
    path: var/db/baetyl/test-hub-conf
  - name: test-timer-conf
    path: var/db/baetyl/test-timer-conf`

	dir, err := ioutil.TempDir("", "template")
	assert.NoError(t, err)
	fileName1 := "compose_conf"
	f, err := os.Create(filepath.Join(dir, fileName1))
	defer f.Close()
	_, err = io.WriteString(f, composeConfString)
	assert.NoError(t, err)
	cfg2, err := LoadComposeAppConfigCompatible(filepath.Join(dir, fileName1))
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg2)

	fileName2 := "conf"
	f2, err := os.Create(filepath.Join(dir, fileName2))
	defer f2.Close()
	_, err = io.WriteString(f2, confString)
	assert.NoError(t, err)
	cfg3, err := LoadComposeAppConfigCompatible(filepath.Join(dir, fileName2))
	assert.NoError(t, err)
	assert.Equal(t, cfg, cfg3)
}
