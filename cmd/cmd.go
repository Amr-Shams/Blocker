package cmd

import (
	"fmt"
	"time"

	"github.com/Amr-Shams/Blocker/blockchain"
	n "github.com/Amr-Shams/Blocker/node"
	"github.com/Amr-Shams/Blocker/wallet"
	"github.com/pkg/profile"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// TODO(66): Add the auto-completion for the commands
// it is a good way as there are a lot of commands

var (
	avilaibleNodeCommands   = []string{"print", "create", "clean", "refresh"}
	avilaibleWalletCommands = []string{"createwallet", "listalladdress", "clean", "inquirebalance", "send"}
)

func addCommandsToNode(node *cobra.Command) {
	node.AddCommand(blockchain.PrintCommand())
	node.AddCommand(blockchain.CreateBlockChainCommand())
	node.AddCommand(blockchain.CleanCommand())
	node.AddCommand(blockchain.ReindexCommand())
}

func addCommandsToWallet(w *cobra.Command) {
	w.AddCommand(wallet.ListALLAdressCommand())
	w.AddCommand(wallet.CreateWalletCommand())
	w.AddCommand(wallet.InquireyBalanceCommand())
	w.AddCommand(wallet.SendCommand())
	w.AddCommand(wallet.PrintCommand())
}

type prof interface {
	Stop()
}

func NewNodeCommand() *cobra.Command {
	result := &cobra.Command{
		Use:     "node",
		Short:   "Blockchain Node",
		Long:    "A blockchain node implementation in Go",
		Example: "node print",
		Args:    cobra.ExactArgs(0),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, unusedCommands := lo.Difference(avilaibleNodeCommands, args)
			return unusedCommands, cobra.ShellCompDirectiveNoFileComp
		},
		ValidArgs: avilaibleNodeCommands,
	}
	addCommandsToNode(result)
	return result
}

func NewWalletCommand() *cobra.Command {

	result := &cobra.Command{
		Use:     "wallet",
		Short:   "Blockchain Wallet",
		Long:    "A blockchain wallet implementation in Go",
		Example: "wallet create",
		Args:    cobra.ExactArgs(0),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, unUsed := lo.Difference(avilaibleWalletCommands, args)
			return unUsed, cobra.ShellCompDirectiveNoFileComp
		},
		ValidArgs: avilaibleWalletCommands,
	}

	addCommandsToWallet(result)
	return result
}

func NewRootCommand() *cobra.Command {
	var (
		start    time.Time
		profiler prof
	)
	root := &cobra.Command{
		Short:   "Blocker is a blockchain implementation in Go",
		Example: `blocker node print -n localhost:3000`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			start = time.Now()
			if viper.GetBool("profile") {
				profiler = profile.Start(profile.CPUProfile, profile.ProfilePath("."))
			}
			if viper.GetString("NodeID") == "" {
				panic("NodeID is required for the blockchain")
			}
			viper.Set("NodeID", fmt.Sprintf("localhost:%s", viper.GetString("NodeID")))
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if viper.GetBool("profile") {
				profiler.Stop()
			}
			fmt.Printf("Time taken: %v\n", time.Since(start))
		},
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			_, unusedCommands := lo.Difference([]string{"node", "wallet"}, args)
			return unusedCommands, cobra.ShellCompDirectiveNoFileComp
		},
	}

	flags := root.PersistentFlags()

	flags.BoolP("profile", "p", false, "Enable profiling")
	flags.StringP("NodeID", "n", "", "Node ID for the blockchain")
	root.AddCommand(NewNodeCommand())
	root.AddCommand(NewWalletCommand())
	root.AddCommand(n.StartFullNodeCommand())
	root.AddCommand(n.StartWalletNodeCommand())
	viper.AutomaticEnv()
	_ = viper.BindPFlags(flags)

	return root

}
