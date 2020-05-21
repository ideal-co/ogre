package config

import (
	"encoding/json"
	"github.com/lowellmower/ogre/pkg/install"
	"github.com/moogar0880/venom"
	"io/ioutil"
	"os"
	"path/filepath"
)

/*
NOTE: the backend config stanza
  "backends": [
    {
      "type": "statsd",
      "server": "127.0.0.1:8125",
      "prefix": "ogre"
    },
    {
      "type": "http",
      "server": "127.0.0.1:9009",
      "format": "json",
      "resource_path": "/health"
    }
  ]
*/

var DefaultConfigConst = `
{
  "dockerd_socket": "/run/docker.sock",
  "containerd_socket": "/run/containerd/containerd.sock",
  "ogred_socket": "/var/run/ogred.sock",
  "ogred_pid": "/etc/ogre/ogred.pid",
  "log": {
    "level": "trace",
    "file": "/var/log/ogred.log",
    "silent": false,
    "report_caller": false
  },
  "backends": [
    {
      "type": "statsd",
      "server": "127.0.0.1:8125",
      "prefix": "foo-noodle"
    }
  ],
  "services": [
    {
      "type": "docker",
      "log": {
        "level": "info",
        "file": "/var/log/service.log",
        "silent": false,
        "report_caller": false
      }
    }
  ]
}
`

// DaemonConfig is the structural representation of a ogred.conf.json file.
type DaemonConfig struct {
	DockerdSocket    string          `json:"dockerd_socket,omitempty"`
	ContainerdSocket string          `json:"containerd_socket,omitempty"`
	OgredSocket      string          `json:"ogred_socket"`
	OgredPID         string          `json:"ogred_pid"`
	Log              LogConfig       `json:"log"`
	Backends         []BackendConfig `json:"backends,omitempty"`
	Services         []ServiceConfig `json:"services,omitempty"`
}

// LogConfig is the structural representation of the config for the application's
// global Logger instances.
type LogConfig struct {
	Level        string `json:"level"`
	File         string `json:"file"`
	Silent       bool   `json:"silent"`
	ReportCaller bool   `json:"report_caller"`
}

// BackendConfig is the structural representation of all possible fields which
// could be used to configure a backend. Required fields do not have the
// omitempty decorator and must be present.
type BackendConfig struct {
	// shared
	Type   string `json:"type"`
	Server string `json:"server"`

	// statsd
	Prefix string `json:"prefix,omitempty"`

	// prometheus
	Job    string `json:"job,omitempty"`
	Metric string `json:"metric,omitempty"`

	// http(s)
	Format       string `json:"format,omitempty"`
	ResourcePath string `json:"resource_path,omitempty"`
}

// ServiceConfig is the structural representation of a service in the config
// file. At the time of writing this, the only exposed service is docker and
// that service is also enabled by default.
type ServiceConfig struct {
	Type string    `json:"type"`
	Log  LogConfig `json:"log"`
}

// our application wide config values
var Daemon = &venom.Venom{}
var Service = &venom.Venom{}

var DaemonConf = &DaemonConfig{}

// init for the config package will attempt to load config from the default file
// path or, if that does not exist, load it from the string literal of defaults
// above and then write that to said path.
func init() {
	// check if config file exists
	ogredConf := install.HostConfigDir + install.OgredConfig
	if confFile, _ := os.Stat(ogredConf); confFile != nil {
		// if the file exists on host, load that file
		LoadConfig()
		return
	}

	// if no config existed on host, load default and write config
	LoadDefaults()
}

// LoadConfig makes the assumption there is a file in place at the default file
// path for config, /etc/ogre/ogre.d/ogred.conf.json
func LoadConfig() {
	data, err := ioutil.ReadFile(install.HostConfigDir + install.OgredConfig)
	if err != nil {
		panic(err)
	}
	if err = json.Unmarshal(data, DaemonConf); err != nil {
		panic(err)
	}
	Daemon = readConfig(install.HostConfigDir + install.OgredConfig)
	// TODO (lmower): determine use case for separate service config
	//                duplicating daemon config for now.
	Service = readConfig(install.HostConfigDir + install.OgredConfig)
}

// LoadDefaults will parse the literal string above into the DaemonConfig
// struct, write that to a file on the host and then set the
func LoadDefaults() {
	writeFromMemory()
	Daemon = readConfig(install.HostConfigDir + install.OgredConfig)
	// TODO (lmower): determine use case for separate service config
	//                duplicating daemon config for now.
	Service = readConfig(install.HostConfigDir + install.OgredConfig)
}

func writeFromMemory() {
	//var conf = &DaemonConfig{}
	if err := json.Unmarshal([]byte(DefaultConfigConst), DaemonConf); err != nil {
		panic(err)
	}
	hostFile := filepath.Join(install.HostConfigDir, install.OgredConfig)
	if _, err := os.Stat(hostFile); err != nil {
		if err = os.MkdirAll(install.HostConfigDir, os.FileMode(os.O_RDWR)); err != nil {
			panic("config dir did not exist and could not be created: " + err.Error())
		}
		if _, err = os.Create(hostFile); err != nil {
			panic("config file did not exist and could not be created: " + err.Error())
		}
	}
	data, err := json.MarshalIndent(DaemonConf, "", "    ")
	if err != nil {
		panic("could not unmarshal default config: " + err.Error())
	}
	if err = ioutil.WriteFile(hostFile, data, os.FileMode(os.O_RDWR)); err != nil {
		panic("could not write ogre configuration: " + err.Error())
	}
}

func readConfig(config string) *venom.Venom {
	v := venom.Default()
	if err := v.LoadDirectory(config, false); err != nil {
		panic(err)
	}
	return v
}
