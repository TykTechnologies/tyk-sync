package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// examplesCmd represents a command to show and publish examples from an examples repository
var examplesCmd = &cobra.Command{
	Use:   "examples",
	Short: "Shows a list of all available tyk examples",
	Long: `This command will show a list of all available tyk examples that are hosted on the official
	GitHub repository. For more details please use the 'examples show' command.'`,
	Run: func(cmd *cobra.Command, args []string) {
		err := processExamplesList()
		if err != nil {
			fmt.Println("Error: ", err)
			os.Exit(1)
		}
	},
}

func init() {
	RootCmd.AddCommand(examplesCmd)
}
