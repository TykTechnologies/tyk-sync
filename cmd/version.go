package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

const VERSION = "1.2.3"

func init() {
	RootCmd.AddCommand(versionCmd)
}

var versionCmd = &cobra.Command{
	Use:     "version",
	Aliases: []string{"v"},
	Short:   "Tyk-sync version",
	Long:    `This command will show the current Tyk-Sync version.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("v" + VERSION)
	},
}
