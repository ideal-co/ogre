package install

import (
    "path"
    "runtime"
)

// Default directories
const (
    // config
    HostConfigDir = "/etc/ogre/ogre.d/"
    AppConfigDir = "configs/ogre.d/"

    // bin
    HostBinDir = "/usr/local/bin/"
    AppBinDir = "bin/"
)

// Default config files
const (
    DockerConfig = "docker.conf.json"
    OgredConfig = "ogred.conf.json"
)

// Default executable file paths
const (
    OgredBin = "ogred"
)

// misc
const (
    HostPIDFilepath = "/etc/ogre/ogred.pid"
)

// AppRoot is set on the init() below to be the path to the project root so that
// when the project is cloned into any go environment it should, "just work"
var AppRoot string

func init() {
    _, base, _, _ := runtime.Caller(0)
    AppRoot = path.Join(path.Dir(base), "../..") + "/"
}