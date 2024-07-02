package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// syncCmd represents the sync command
var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Synchronise API configurations from version control or file system to Tyk Gateway or Dashboard",
	Long: `This command will synchronise an API Gateway/Dashboard with the contents of a Github repository, the
	sync is one way: from the repo to the gateway/dashboard, the command will not write back to the repo.
	Sync will delete any objects in the dashboard or gateway that it cannot find in the github repo,
	update those that it can find and create those that are missing.`,
	Example: `Synchronise from Git repository:
	tyk-sync sync {-d DASHBOARD_URL | -g GATEWAY_URL} [-s SECRET] [-b BRANCH] [-k SSHKEY] [-o ORG_ID] REPOSITORY_URL

Synchronise from file system:
	tyk-sync sync {-d DASHBOARD_URL | -g GATEWAY_URL} [-s SECRET] [-o ORG_ID] -p PATH`,
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

	syncCmd.Flags().SortFlags = false

	syncCmd.Flags().StringP("gateway", "g", "", "Specify the fully qualified URL of the Tyk Gateway where configuration changes should be applied (Either -d or -g is required)")
	syncCmd.Flags().StringP("dashboard", "d", "", "Specify the fully qualified URL of the Tyk Dashboard where configuration changes should be applied (Either -d or -g is required)")
	syncCmd.Flags().StringP("key", "k", "", "Provide the location of the SSH key file for authentication to Git (optional)")
	syncCmd.Flags().StringP("branch", "b", "refs/heads/master", "Specify the branch of the GitHub repository to use")
	syncCmd.Flags().StringP("secret", "s", "", "API secret for accessing Dashboard or Gateway API (optional).  If not set, value of TYKGIT_DB_SECRET/TYKGIT_GW_SECRET environment variable will be used for dahboard or gateway respectively")
	syncCmd.Flags().StringP("org", "o", "", "Override the organization ID to use for the synchronisation process (optional)")
	syncCmd.Flags().StringP("path", "p", "", "Specify the source file directory where API configuration files are located (Required for synchronising from file system)")
	syncCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
	syncCmd.Flags().Bool("allow-unsafe-oas", false, "Use Tyk Classic endpoints in Tyk Dashboard API for Tyk OAS APIs (optional)")
	syncCmd.Flags().StringSlice("apis", []string{}, "Specify API IDs to synchronise. These APIs will be created or updated during synchronisation. Other resources will be deleted")
	syncCmd.Flags().StringSlice("policies", []string{}, "Specify policy IDs to synchronise. These policies will be created or updated during synchronisation. Other resources will be deleted")

	if err := syncCmd.Flags().MarkDeprecated("allow-unsafe-oas", "OAS API can be synced without the flag."); err != nil {
		fmt.Printf("Failed to mark `allow-unsafe-oas` flag as deprecated: %v", err)
	}

	if err := syncCmd.Flags().MarkDeprecated("apis", "The flag will be removed in upcoming release"); err != nil {
		fmt.Printf("Failed to mark `apis` flag as deprecated: %v", err)
	}

	if err := syncCmd.Flags().MarkDeprecated("policies", "The flag will be removed in upcoming release"); err != nil {
		fmt.Printf("Failed to mark `policies` flag as deprecated: %v", err)
	}
}
