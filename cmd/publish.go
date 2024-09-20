package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish new API configurations to Tyk Gateway or Dashboard",
	Example: `Publish from Git repository:
	tyk-sync publish {-d DASHBOARD_URL | -g GATEWAY_URL} [-s SECRET] [-b BRANCH] [-k SSHKEY] [-o ORG_ID] REPOSITORY_URL

Publish from file system:
	tyk-sync publish {-d DASHBOARD_URL | -g GATEWAY_URL} [-s SECRET] [-o ORG_ID] -p PATH
	`,
	Long: `Publish API definitions from a Git repo to a gateway or dashboard, this
	will not update existing APIs, and if it detects a collision, will stop.`,
	Run: func(cmd *cobra.Command, args []string) {
		verificationError := verifyArguments(cmd)
		if verificationError != nil {
			fmt.Println(verificationError)
			os.Exit(1)
		}

		err := processPublish(cmd, args)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(publishCmd)

	publishCmd.Flags().SortFlags = false

	// Here you will define your flags and configuration settings.
	publishCmd.Flags().StringP("gateway", "g", "", "Specify the fully qualified URL of the Tyk Gateway where configuration changes should be applied (Either -d or -g is required)")
	publishCmd.Flags().StringP("dashboard", "d", "", "Specify the fully qualified URL of the Tyk Dashboard where configuration changes should be applied (Either -d or -g is required)")
	publishCmd.Flags().StringP("key", "k", "", "Provide the location of the SSH key file for authentication to Git (optional)")
	publishCmd.Flags().StringP("branch", "b", "refs/heads/master", "Specify the branch of the GitHub repository to use")
	publishCmd.Flags().StringP("secret", "s", "", "API secret for accessing Dashboard or Gateway API (optional).  If not set, value of TYKGIT_DB_SECRET/TYKGIT_GW_SECRET environment variable will be used for dahboard or gateway respectively")
	publishCmd.Flags().StringP("org", "o", "", "Override the organization ID to use for the synchronisation process (optional)")
	publishCmd.Flags().StringP("path", "p", "", "Specify the source file directory where API configuration files are located (Required for synchronising from file system)")
	publishCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
	publishCmd.Flags().Bool("allow-unsafe-oas", false, "Use Tyk Classic endpoints in Tyk Dashboard API for Tyk OAS APIs (optional)")
	publishCmd.Flags().StringSlice("apis", []string{}, "Specify API IDs to publish. Use this to selectively publish specific APIs. It can be a single ID or an array of string such as “id1,id2”")
	publishCmd.Flags().StringSlice("oas-apis", []string{}, "Specify OAS API IDs to dump. Use this to selectively dump specific OAS APIs. It can be a single ID or an array of string such as “id1,id2”")
	publishCmd.Flags().StringSlice("policies", []string{}, "Specify policy IDs to publish. Use this to selectively publish specific policies. It can be a single ID or an array of string such as “id1,id2”")
	publishCmd.Flags().StringSlice("templates", []string{}, "Specify template IDs to publish. Use this to selectively publish specific API templates. It can be a single ID or an array of string such as “id1,id2”")

	if err := publishCmd.Flags().MarkDeprecated("allow-unsafe-oas", "OAS API can published without the flag."); err != nil {
		fmt.Printf("Failed to mark `allow-unsafe-oas` flag as deprecated: %v", err)
	}
}
