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
	Short: "Update a dashboard or gateway with APIs and policies",
	Long: `Update will attempt to identify matching APIs or Policies in the target, and update those APIs
	It will not create new ones, to do this use publish or sync.`,
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
	updateCmd.Flags().StringP("gateway", "g", "", "Fully qualified gateway target URL")
	updateCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	updateCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	updateCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	updateCmd.Flags().StringP("secret", "s", "", "Your API secret")
	updateCmd.Flags().StringP("path", "p", "", "Source directory for definition files (optional)")
	updateCmd.Flags().Bool("test", false, "Use test publisher, output results to stdio")
}
