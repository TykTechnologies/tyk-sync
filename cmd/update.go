// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// updateCmd represents the update command
var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update existing API configurations in Tyk Gateway or Dashboard",
	Long: `Update will attempt to identify matching APIs or Policies in the target, and update those APIs
	It will not create new ones, to do this use publish or sync.`,
	Example: `Update from Git repository:
	tyk-sync update {-d DASHBOARD_URL | -g GATEWAY_URL} [-s SECRET] [-b BRANCH] [-k SSHKEY] [-o ORG_ID] REPOSITORY_URL

Update from file system:
	tyk-sync update {-d DASHBOARD_URL | -g GATEWAY_URL} [-s SECRET] [-o ORG_ID] -p PATH`,
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
	RootCmd.AddCommand(updateCmd)
	updateCmd.Flags().SortFlags = false

	updateCmd.Flags().StringP("gateway", "g", "", "Specify the fully qualified URL of the Tyk Gateway where configuration changes should be applied (Either -d or -g is required)")
	updateCmd.Flags().StringP("dashboard", "d", "", "Specify the fully qualified URL of the Tyk Dashboard where configuration changes should be applied (Either -d or -g is required)")
	updateCmd.Flags().StringP("key", "k", "", "Provide the location of the SSH key file for authentication to Git (optional)")
	updateCmd.Flags().StringP("branch", "b", "refs/heads/master", "Specify the branch of the GitHub repository to use")
	updateCmd.Flags().StringP("secret", "s", "", "Your API secret for accessing Dashboard or Gateway API (optional)")
	updateCmd.Flags().StringP("path", "p", "", "Specify the source file directory where API configuration files are located (Required for synchronising from file system)")
	updateCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
	updateCmd.Flags().Bool("allow-unsafe-oas", false, "Use Tyk Classic endpoints in Tyk Dashboard API for Tyk OAS APIs (optional)")
	updateCmd.Flags().StringSlice("apis", []string{}, "Specify API IDs to update. Use this to selectively update specific APIs. It can be a single ID or an array of string such as “id1,id2”")
	updateCmd.Flags().StringSlice("oas-apis", []string{}, "Specify OAS API IDs to dump. Use this to selectively dump specific OAS APIs. It can be a single ID or an array of string such as “id1,id2”")
	updateCmd.Flags().StringSlice("policies", []string{}, "Specify policy IDs to update. Use this to selectively update specific policies. It can be a single ID or an array of string such as “id1,id2”")
	updateCmd.Flags().StringSlice("templates", []string{}, "Specify template IDs to update. Use this to selectively update specific API templates. It can be a single ID or an array of string such as “id1,id2”")

	if err := updateCmd.Flags().MarkDeprecated("allow-unsafe-oas", "OAS API can updated without the flag."); err != nil {
		fmt.Printf("Failed to mark `allow-unsafe-oas` flag as deprecated: %v", err)
	}
}
