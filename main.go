package main

import (
	"github.com/Amr-Shams/Blocker/rootCmd"
)

func main() {
	cmd := rootCmd.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
