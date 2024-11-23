package rootCmd

import (
	cmd "github.com/Amr-Shams/Blocker/cmd"
	n "github.com/Amr-Shams/Blocker/node"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	root := cmd.BaseCommand()
	root.AddCommand(n.StartFullNodeCommand())
	root.AddCommand(n.StartWalletNodeCommand())
	return root
}
