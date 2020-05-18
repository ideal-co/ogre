package config

import (
	"github.com/lowellmower/ogre/pkg/install"
	"github.com/moogar0880/venom"
	"io/ioutil"
	"os"
	"path/filepath"
)

var Daemon = &venom.Venom{}
var Service = &venom.Venom{}

func init() {
	writeDefaultConfig()
	Daemon = readConfig(install.HostConfigDir + install.OgredConfig)
	Service = readConfig(install.HostConfigDir + install.ServiceConfig)
}

func writeDefaultConfig() {
	confDir := install.AppRoot + install.AppConfigDir
	files, err := ioutil.ReadDir(confDir)
	if err != nil {
		panic("could not read ogre configuration: " + err.Error())
	}
	for _, f := range files {
		data, err := ioutil.ReadFile(confDir + f.Name())
		if err != nil {
			panic("could not read ogre configuration: " + err.Error())
		}
		hostFile := filepath.Join(install.HostConfigDir, f.Name())
		if _, err := os.Stat(hostFile); err != nil {
			if err = os.MkdirAll(install.HostConfigDir, os.FileMode(os.O_RDWR)); err != nil {
				panic("config dir did not exist and could not be created: " + err.Error())
			}
			if _, err = os.Create(hostFile); err != nil {
				panic("config file did not exist and could not be created: " + err.Error())
			}
		}
		if err = ioutil.WriteFile(hostFile, data, os.FileMode(os.O_RDWR)); err != nil {
			panic("could not read ogre configuration: " + err.Error())
		}
	}
}

func readConfig(config string) *venom.Venom {
	v := venom.Default()
	if err := v.LoadDirectory(config, false); err != nil {
		panic(err)
	}
	return v
}
