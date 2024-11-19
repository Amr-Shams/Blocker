package blockchain

import (
	"fmt"
	"log"
	"runtime"

	"github.com/Amr-Shams/Blocker/storage"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func PrintCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "print",
		Short: "Prints blockchain",
		Run: func(_ *cobra.Command, _ []string) {
			nodeID := viper.GetString("NodeID")
			chain := ContinueBlockChain(nodeID)
			chain.Print()
		},
	}
}

func CreateBlockChainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new blockchain",
		Run: func(_ *cobra.Command, _ []string) {
			address := viper.GetString("wallet")
			if !util.ValidateAddress(address) {
				log.Println("Invalid address")
				runtime.Goexit()
			}
			nodeID := viper.GetString("NodeID")
			bc := NewBlockChain(address, nodeID)
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			defer bc.Close()
			fmt.Println("Blockchain created")
		},
	}
	flags := cmd.PersistentFlags()
	flags.StringP("wallet", "w", "", "Address to Create BlockChain")
	cmd.MarkPersistentFlagRequired("address")
	viper.BindPFlag("wallet", flags.Lookup("wallet"))
	return cmd
}
func CleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean the blockchain",
		Run: func(_ *cobra.Command, _ []string) {
			nodeID := viper.GetString("NodeID")
			db, _ := storage.GetInstance(nodeID)
			defer db.Clean()
		},
	}
}

func ReindexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Reindex the UTXO set",
		Run: func(_ *cobra.Command, _ []string) {
			nodeID := viper.GetString("NodeID")
			bc := ContinueBlockChain(nodeID)
			defer bc.Close()
			UTXOSet := UTXOSet{bc}
			UTXOSet.Reindex()
			count := UTXOSet.CountTransactions()
			fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
		},
	}
}
