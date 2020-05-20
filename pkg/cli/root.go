package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ogre",
	Short: "Container monitoring for modern distributed systems.",
	Long: `Ogre provides a zero configuration service which can immediately be
used to begin monitoring containers. It is designed for integration into modern
day distibuted systems and the surrounding Docker eco-system.

Ogre can be deployed on bare metal or in a container and will by default begin
offering insight to the eco-system's health. Ogre is also configurable and can
be integrated into roll-your-own solutions and other Docker services such as
Kubernetes and Swarm.'`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//      Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Run() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize()
}
