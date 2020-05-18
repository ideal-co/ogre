package cli

import (
	"fmt"
	"github.com/spf13/cobra"
	"strings"
)

// add all commands to root command in cmd/ogre/root.go
func init() {
	configCmd.AddCommand(listSubCmd, setSubCmd)
	rootCmd.AddCommand(configCmd)
}

/*
Below are all the sub commands of ogre related to configurations and are to
be accessed under command structure like: ogre config [sub-cmd] [flags]
*/

// top level config, ogre config
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Set and retrieve information about Ogre's configuration.",
	Long: `Ogre can be configured to do a number of things, this top level
command is used to set or retrieve configuration values at runtime or before
the Ogre daemon is started.'`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("top level config")
	},
}

// TODO (lmower): read config from disk and print to stdout
// sub command list, ogre config list
var listSubCmd = &cobra.Command{
	Use:   "list",
	Short: "Print out the configuration values currently set.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("printing config list...")
	},
}

// TODO (lmower): actually have this set some value and persist state
// sub command set, ogre config set [args]
var setSubCmd = &cobra.Command{
	Use:   "set",
	Short: "Set or replace a value in the configuration.",
	Args:  cobra.MinimumNArgs(1),
	Long: `Setting or replacing a value WILL NOT be recognized by the daemon
until the process is restarted or configuration is reloaded. Use key value pairs
associated by an equal sign and separated by a space if there are multiple.`,
	Example: "config set dockerd_socket=/var/run/docker.sock dhcp_ports=68,67",
	Run: func(cmd *cobra.Command, args []string) {
		for _, arg := range args {
			kv := strings.Split(arg, "=")
			if len(kv) != 2 {
				continue
			}
			fmt.Printf("Key: %s, Val: %s\n", kv[0], kv[1])
		}
	},
}
