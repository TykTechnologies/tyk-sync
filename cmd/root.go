package cmd

import "github.com/spf13/cobra"

func init() {

}

var RootCmd = &cobra.Command{
	Use:   "tyk-sync",
	Short: "Tyk Git is a tool to integrate Tyk Gateway with Git",
	Long: `Tyk Sync lets you store API definitions, security policies, and API templates as files in version control system (VCS) 
or file system and synchronise changes to Tyk, promoting a consistent and automated approach to managing API configurations.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
