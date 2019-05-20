package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronise a github repo or file system with a gateway",
	Long: `This command will synchronise an API Gateway with the contents of a Github repository, the
	sync is one way: from the repo to the gateway, the command will not write back to the repo.
	Sync will delete any objects in the dashboard or gateway that it cannot find in the github repo,
	update those that it can find and create those that are missing.`,
	Run: func(cmd *cobra.Command, args []string) {
		verificationError := verifyArguments(cmd)
		if verificationError != nil {
			fmt.Println(verificationError)
			os.Exit(1)
		}

		err := processSync(cmd, args)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(syncCmd)

	syncCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	syncCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	syncCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	syncCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	syncCmd.Flags().StringP("secret", "s", "", "Your API secret")
	syncCmd.Flags().StringP("org", "o", "", "org ID override")
	syncCmd.Flags().StringP("path", "p", "", "Source directory for definition files (optional)")
	syncCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
}
