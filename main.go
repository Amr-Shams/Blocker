package main

import (
	"github.com/Amr-Shams/Blocker/cmd"
)

func main() {
	// 
	
	cmd := cmd.NewRootCommand()
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
}
