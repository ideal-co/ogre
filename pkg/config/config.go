package config

import (
	"github.com/lowellmower/ogre/pkg/install"
	"github.com/moogar0880/venom"
	"io/ioutil"
	"os"
)

var Daemon = &venom.Venom{}

func init() {
	writeDefaultConfig()
	Daemon = readConfig(install.HostConfigDir + install.OgredConfig)
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
		if err = ioutil.WriteFile(install.HostConfigDir + f.Name(), data, os.FileMode(os.O_RDWR)); err != nil {
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
