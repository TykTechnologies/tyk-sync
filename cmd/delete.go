package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete API and/or policy definitions from a gateway or dashboard",
	Run: func(cmd *cobra.Command, args []string) {
		verificationError := verifyArguments(cmd)
		if verificationError != nil {
			fmt.Println(verificationError)
			os.Exit(1)
		}
		apis, apisFlagError := cmd.Flags().GetStringSlice("apis")
		if apisFlagError != nil {
			fmt.Printf("Error with the specified apis: %v\n", apisFlagError)
			os.Exit(1)
		}
		pols, polsFlagError := cmd.Flags().GetStringSlice("policies")
		if polsFlagError != nil {
			fmt.Printf("Error with the specified policies: %+v\n", polsFlagError)
			os.Exit(1)
		}
		if len(apis) == 0 && len(pols) == 0 {
			fmt.Printf("Error: please specify --apis and/or --policies to delete\n\n")
			cmd.Help()
			os.Exit(1)
		}

		err := processDelete(cmd, args)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	deleteCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	deleteCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	deleteCmd.Flags().StringP("secret", "s", "", "Your API secret")
	deleteCmd.Flags().Bool("test", false, "Use test mode, output results to stdio")
	deleteCmd.Flags().StringSlice("policies", []string{}, "Specific Policies ids to delete")
	deleteCmd.Flags().StringSlice("apis", []string{}, "Specific Apis ids to delete")
}
