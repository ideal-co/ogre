package cli

import (
	"fmt"
	"github.com/lowellmower/ogre/pkg/config"
	"github.com/lowellmower/ogre/pkg/daemon"
	"github.com/spf13/cobra"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// add all commands to root command in cmd/ogre/root.go
func init() {
	rootCmd.AddCommand(stopCmd)
}

// stopCmd stops the ogred process and removes files associated with the running
// proc such as the socket file descriptor and the PID file.
var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stops the ogre daemon",
	RunE: func(cmd *cobra.Command, args []string) error {
		ogredSock := config.Daemon.GetString(daemon.OgredSocket)
		ogredPID := config.Daemon.GetString(daemon.OgredPIDFile)

		defer func() {
			data, err := ioutil.ReadFile(ogredPID)
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Printf("stopping process: %s", string(data))
			pid, err := strconv.Atoi(strings.Trim(string(data), "\n"))
			if err != nil {
				fmt.Println(err)
			}

			err = syscall.Kill(pid, syscall.SIGKILL)
			if err != nil {
				fmt.Println(err)
			}

			os.RemoveAll(ogredPID)
		}()

		if len(ogredSock) == 0 {
			return fmt.Errorf("no unix socket configured by key %s", daemon.OgredSocket)
		}
		daemonClient, err := net.Dial("unix", ogredSock)
		if err != nil {
			return fmt.Errorf("could not connect client to unix socket %s: %s", ogredSock, err)
		}
		if _, err = daemonClient.Write([]byte(`{"service": "daemon", "action": "stop"}`)); err != nil {
			return fmt.Errorf("error sending stop command to daemon: %s", err)
		}

		return nil
	},
}
