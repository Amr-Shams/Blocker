package main

import (
	"github.com/Amr-Shams/Blocker/root"
)

func main() {
	cmd := root.NewRootCommand()

	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
