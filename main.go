package main

import (
	"github.com/Amr-Shams/Blocker/cmd"
)

func main() {
	// TODO: cmd update the pointer to the file
	// use a singleton pattern to avoid multiple instances for the db(filesystem)
	cmd := cmd.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
