package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// examplesPublishCmd represents a command to publish a specific example to a gateway or dashboard
var examplesPublishCmd = &cobra.Command{
	Use:   "publish",
	Short: "Publish a specific example to a gateway or dashboard by using its location",
	Long: `This command will publish a specific example by providing the location to a gateway or dashboard. The location can be
	determined by using the 'examples' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		verificationError := verifyArguments(cmd)
		if verificationError != nil {
			fmt.Println(verificationError)
			os.Exit(1)
		}

		err := processExamplePublish(cmd, args)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	examplesCmd.AddCommand(examplesPublishCmd)

	examplesPublishCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	examplesPublishCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	examplesPublishCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	examplesPublishCmd.Flags().StringP("branch", "b", "refs/heads/main", "Branch to use (defaults to refs/heads/main)")
	examplesPublishCmd.Flags().StringP("secret", "s", "", "Your API secret")
	examplesPublishCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
	examplesPublishCmd.Flags().StringP("location", "l", "", "Location to example")
	err := examplesPublishCmd.MarkFlagRequired("location")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
