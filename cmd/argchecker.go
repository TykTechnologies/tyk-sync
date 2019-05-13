package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func verifyArguments(cmd *cobra.Command) error {
	gwString, _ := cmd.Flags().GetString("gateway")
	dbString, _ := cmd.Flags().GetString("dashboard")
	brString, _ := cmd.Flags().GetString("branch")
	ptString, _ := cmd.Flags().GetString("path")

	if gwString == "" && dbString == "" {
		return errors.New(fmt.Sprintf("%s requires either gateway or dashboard target to be set", cmd.Use))
	}

	if gwString != "" && dbString != "" {
		return errors.New(fmt.Sprintf("%s requires either gateway or dashboard target to be set, not both", cmd.Use))
	}

	if ptString != "" && brString != "refs/heads/master" {
		return errors.New(fmt.Sprintf("%s requires either files or branch to be set, not both", cmd.Use))
	}
	return nil
}

