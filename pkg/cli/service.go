package cli

import (
	"fmt"
	"github.com/lowellmower/ogre/pkg/config"
	"github.com/lowellmower/ogre/pkg/daemon"
	"github.com/lowellmower/ogre/pkg/log"
	"github.com/spf13/cobra"
	"net"
	"os"
)

// add all commands to root command in cmd/ogre/root.go
func init() {
	serviceCmd.AddCommand(serviceSubCmd)
	rootCmd.AddCommand(serviceCmd)
}

var serviceCmd = &cobra.Command{
	Use:   "service",
	Short: "Interface with ogre services",
	Long: `Service allows you to inspect services that may or may not be running as
well as control and configure them.`,
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ogrePID := config.Daemon.GetString(daemon.OgredPIDFile)

		// check if daemon is already running
		if _, err := os.Stat(ogrePID); err != nil {
			log.Daemon.Error("ogre daemon is not running, run: ogre start")
			return fmt.Errorf("ogre daemon is not running, run: ogre start")
		}

		return nil
	},
}

var serviceSubCmd = &cobra.Command{
	Use:   "docker",
	Short: "Send a command to the ogre docker service",
	Long: `The docker service sub-command allows you to send signals to the docker
API via the ogre listener. You can start, stop, inspect, enable, and disable the
docker service functions by this interface.`,
	Args:    cobra.MinimumNArgs(1),
	Example: "ogre service docker start",
	RunE: func(cmd *cobra.Command, args []string) error {
		sock := config.Daemon.GetString(daemon.OgredSocket)
		if len(sock) == 0 {
			return fmt.Errorf("no unix socket configured by key %s", daemon.OgredSocket)
		}
		daemonClient, err := net.Dial("unix", sock)
		if err != nil {
			return fmt.Errorf("could not connect client to unix socket %s: %s", sock, err)
		}

		switch args[0] {
		case "start":
			daemonClient.Write([]byte(`{"service": "docker", "action": "start"}`))
		case "stop":
			daemonClient.Write([]byte(`{"service": "docker", "action": "stop"}`))
		}

		return nil
	},
}
