package cli

import (
	"fmt"
	"github.com/ideal-co/ogre/pkg/config"
	"github.com/ideal-co/ogre/pkg/daemon"
	"github.com/ideal-co/ogre/pkg/install"
	"github.com/spf13/cobra"
	"os"
	"strconv"
)

// add all commands to root command in cmd/ogre/root.go
func init() {
	rootCmd.AddCommand(startCmd)
}

// startCmd starts the ogred process by way of executing the ogred binary
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts the ogre daemon",
	Long: `Starts the ogre daemon which should install in the default location
unless configured otherwise. If a fatal error is signaled from the daemon, it
will be returned to this command as well as the daemon log file.

Note: ensure you have made any custom configurations before running start.'`,
	RunE: func(cmd *cobra.Command, args []string) error {
		var ogredPIDFile string
		var ogreBin string
		// check if daemon is already running using possible config value
		if ogrePID, ok := config.Daemon.Find(daemon.OgredPIDFile); ok {
			ogredPIDFile = ogrePID.(string)
			if msg, exist := ogrePIDFileExist(ogrePID.(string)); exist {
				fmt.Printf("ogred already running %s", msg)
				return nil
			}
		} else {
			// otherwise use default values
			ogredPIDFile = install.HostPIDFilepath
			if msg, exist := ogrePIDFileExist(install.HostPIDFilepath); exist {
				fmt.Printf("ogred already running %s", msg)
				return nil
			}
		}

		argv := make([]string, len(args)+1)
		var proc *os.Process

		if len(config.DaemonConf.OgredBin) != 0 {
			ogreBin = config.DaemonConf.OgredBin + install.OgredBin
		} else {
			ogreBin = install.HostBinDir + install.OgredBin
		}
		argv[0] = ogreBin
		proc, err := os.StartProcess(ogreBin, argv, new(os.ProcAttr))
		if err != nil {
			fmt.Println("Err: ", err)
			return err
		}

		f, err := os.Create(ogredPIDFile)
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

func ogrePIDFileExist(file string) (string, bool) {
	if pidFile, _ := os.Stat(file); pidFile != nil {
		return fmt.Sprintf("since %s, PID at: %s\n", pidFile.ModTime(), file), true
	}
	return "", false
}
