package main

import (
	"fmt"
	"github.com/dmayo3/tyk-sync/cmd"
	"os"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
