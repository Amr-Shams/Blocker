package main

import (
	"os"

	"github.com/Amr-Shams/Blocker/cmd"
)

func main() {
    cmd := cmd.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}

}
