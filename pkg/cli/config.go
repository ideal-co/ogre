package cli

import (
	"encoding/json"
	"fmt"
	"github.com/ideal-co/ogre/pkg/config"
	"github.com/ideal-co/ogre/pkg/install"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
)

// add all commands to root command in cmd/ogre/root.go
func init() {
	configCmd.AddCommand(listSubCmd)
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

// sub command list, ogre config list
var listSubCmd = &cobra.Command{
	Use:   "list",
	Short: "Print out the configuration values currently set.",
	Run: func(cmd *cobra.Command, args []string) {
		conf := &config.DaemonConfig{}
		configPath := install.HostConfigDir + install.OgredConfig
		if info, _ := os.Stat(configPath); info != nil {
			data, err := ioutil.ReadFile(configPath)
			if err != nil {
				panic(err)
			}
			err = json.Unmarshal(data, conf)
			if err != nil {
				panic(err)
			}
			pretty, err := json.MarshalIndent(conf, "", "    ")
			if err != nil {
				panic(err)
			}
			fmt.Println(string(pretty))
			return
		}
		fmt.Printf("no config at %s, run 'ogre start' to get started", configPath)
	},
}

//// TODO (lmower): actually have this set some value and persist state
//// sub command set, ogre config set [args]
//var setSubCmd = &cobra.Command{
//	Use:   "set",
//	Short: "Set or replace a value in the configuration.",
//	Args:  cobra.MinimumNArgs(1),
//	Long: `Setting or replacing a value WILL NOT be recognized by the daemon
//until the process is restarted or configuration is reloaded. Use key value pairs
//associated by an equal sign and separated by a space if there are multiple.`,
//	Example: "config set dockerd_socket=/var/run/docker.sock dhcp_ports=68,67",
//	Run: func(cmd *cobra.Command, args []string) {
//		for _, arg := range args {
//			kv := strings.Split(arg, "=")
//			if len(kv) != 2 {
//				continue
//			}
//			fmt.Printf("Key: %s, Val: %s\n", kv[0], kv[1])
//		}
//	},
//}
