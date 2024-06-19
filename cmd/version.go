package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var VERSION = "1.4.2"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Print version information",
	Long:    `This command will show the current Tyk-Sync version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v" + VERSION)
	},
}
