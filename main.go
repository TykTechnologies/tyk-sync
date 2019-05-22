package main

import (
	"fmt"
	"github.com/TykTechnologies/tyk-sync/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
