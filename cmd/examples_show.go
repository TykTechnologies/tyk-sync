package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// examplesShowCmd represents a command to show the details of a specific example
var examplesShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Shows details of a specific example by using its location",
	Long: `This command will show the details of a specific example by providing the location. The location can be
	determined by using the 'examples' command.`,
	Run: func(cmd *cobra.Command, args []string) {
		err := processExampleDetails(cmd)
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	examplesCmd.AddCommand(examplesShowCmd)

	examplesShowCmd.Flags().StringP("location", "l", "", "Location to example")
	err := examplesShowCmd.MarkFlagRequired("location")
	if err != nil {
		fmt.Println("Error: ", err)
		os.Exit(1)
	}
}
