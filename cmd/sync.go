package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronise a github repo with a gateway",
	Long: `This command will synchronise an API Gateway with the contents of a Github repository, the
	synnc is one way: from the repo to the gateway, the command will not write back to the repo.`,
	Run: func(cmd *cobra.Command, args []string) {
		gwString, _ := cmd.Flags().GetString("gateway")
		dbString, _ := cmd.Flags().GetString("dashboard")

		if gwString == "" && dbString == "" {
			fmt.Println("Sync requires either gateway or dashboard target to be set")
			return
		}

		if gwString != "" && dbString != "" {
			fmt.Println("Sync requires either gateway or dashboard target to be set, not both")
			return
		}

		err := processSync(cmd, args)
		if err != nil {
			fmt.Println("Error: ", err)
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)

	publishCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	publishCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	publishCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	publishCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	publishCmd.Flags().StringP("secret", "s", "", "Your API secret")
	publishCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
}
