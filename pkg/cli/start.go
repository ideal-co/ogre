package cli

import (
	"fmt"
	"github.com/lowellmower/ogre/pkg/config"
	"github.com/lowellmower/ogre/pkg/daemon"
	"github.com/lowellmower/ogre/pkg/install"
	"github.com/lowellmower/ogre/pkg/log"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// add all commands to root command in cmd/ogre/root.go
func init() {
	rootCmd.AddCommand(startCmd)
}

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the ogre daemon",
	Long: `Starts the ogre daemon which should install in the default location
unless configured otherwise. If a fatal error is signaled from the daemon, it
will be returned to this command as well as the daemon log file.

Note: ensure you have made any custom configurations before running start.'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ogrePID := config.Daemon.GetString(daemon.OgredPIDFile)

		// check if daemon is already running
		if pidFile, _ := os.Stat(ogrePID); pidFile != nil {
			data, _ := ioutil.ReadFile(pidFile.Name())
			pid := strings.Trim(string(data), "\n")
			log.Daemon.Infof("ogre daemon running since %s with PID: %s", pidFile.ModTime(), pid)
			return fmt.Errorf("ogre daemon running since %s with PID: %s", pidFile.ModTime(), pid)
		}

		argv := make([]string, len(args)+1)
		var proc *os.Process

		ogreBin := install.AppRoot + install.AppBinDir + install.OgredBin
		argv[0] = ogreBin
		proc, err := os.StartProcess(ogreBin, argv, new(os.ProcAttr))
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		f, err := os.Create(ogrePID)
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}
		defer f.Close()

		if _, err = f.WriteString(strconv.Itoa(proc.Pid) + "\n"); err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		if err = proc.Release(); err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		return nil
	},
}
