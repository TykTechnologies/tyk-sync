package cmd

import "github.com/spf13/cobra"

func init() {

}

var RootCmd = &cobra.Command{
	Use:   "tyk-sync",
	Short: "Tyk Git is a tool to integrate Tyk Gateway with Git",
	Long: `A tool to use Tyk API Definitions or OAS (Swagger) files stored
		in Git (or potentially other VCS) with the Tyk API Management
		Platform (https://tyk.io)`,
	Run: func(cmd *cobra.Command, args []string) {
		// Do Stuff Here
	},
}
