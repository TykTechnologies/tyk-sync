package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// publishCmd represents the publish command
var publishCmd = &cobra.Command{
	Use:   "publish",
	Short: "publish API definitions from a Git repo or file system to a gateway or dashboard",
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

	// Here you will define your flags and configuration settings.
	publishCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	publishCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	publishCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	publishCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	publishCmd.Flags().StringP("secret", "s", "", "Your API secret")
	publishCmd.Flags().StringP("path", "p", "", "Source directory for definition files (optional)")
	publishCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
	publishCmd.Flags().Bool("allow-unsafe-oas", false, "Use Tyk Classic endpoints in Tyk Dashboard API for Tyk OAS APIs (optional)")
	publishCmd.Flags().StringSlice("policies", []string{}, "Specific Policies ids to publish")
	publishCmd.Flags().StringSlice("apis", []string{}, "Specific Apis ids to publish")
	publishCmd.Flags().StringSlice("templates", []string{}, "List of template or Assets IDs to publish (optional")

	if err := publishCmd.Flags().MarkDeprecated("allow-unsafe-oas", "OAS API can published without the flag."); err != nil {
		fmt.Printf("Failed to mark `allow-unsafe-oas` flag as deprecated: %v", err)
	}
}
