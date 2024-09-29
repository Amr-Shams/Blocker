package wallet

import (
	"fmt"
	"log"
	"runtime"

	"github.com/Amr-Shams/Blocker/blockchain"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func ListALLAdressCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lists all addresses from the wallets file",
		Run: func(cmd *cobra.Command, args []string) {
			nodeId := viper.GetString("NodeID")
			GetAllAddresses(nodeId)
		},
	}
}
func CreateWalletCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "add",
		Short: "Creates a new wallet",
		Run: func(cmd *cobra.Command, args []string) {
			nodeId := viper.GetString("NodeID")
			ws := CreateWallets(nodeId)
			ws.AddWallet(nodeId)
		},
	}
}
func SendCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send coins",
		Run: func(_ *cobra.Command, _ []string) {
			from := viper.GetString("from")
			if !util.ValidateAddress(from) {
				log.Println("Invalid address")
				runtime.Goexit()
			}
			to := viper.GetString("to")
			if !util.ValidateAddress(to) {
				log.Println("Invalid address")
				runtime.Goexit()
			}
			amount := viper.GetInt("amount")
			nodeID := viper.GetString("NodeID")
			bc := blockchain.ContinueBlockChain(nodeID)
			ws := LoadWallets(nodeID)
			defer bc.Close()
			UTXO := blockchain.UTXOSet{bc}
			tx := ws.NewTransaction(from, to, amount, &UTXO, nodeID)
			block, added := bc.AddBlock([]*blockchain.Transaction{tx})
			if !added {
				fmt.Println("Failed to add block")
				return
			}
			UTXO.Update(block)
			fmt.Println("Success!")
		},
	}
	flags := cmd.PersistentFlags()
	flags.StringP("from", "f", "", "Address to send from")
	flags.StringP("to", "t", "", "Address to send to")
	flags.StringP("amount", "m", "", "Amount to send")
	cmd.MarkPersistentFlagRequired("from")
	cmd.MarkPersistentFlagRequired("to")
	cmd.MarkPersistentFlagRequired("amount")
	viper.BindPFlag("from", flags.Lookup("from"))
	viper.BindPFlag("to", flags.Lookup("to"))
	viper.BindPFlag("amount", flags.Lookup("amount"))
	return cmd

}
func CleanCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "clean",
		Short: "Clean the wallets file",
		Run: func(cmd *cobra.Command, args []string) {
			nodeId := viper.GetString("NodeID")
			CleanWallet(nodeId)
		},
	}
}
func InquireyBalanceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "Inquirey balance",
		Run: func(_ *cobra.Command, _ []string) {
			address := viper.GetString("address")
			if !util.ValidateAddress(address) {
				log.Println("Invalid address")
				runtime.Goexit()
			}
			nodeID := viper.GetString("NodeID")
			bc := blockchain.ContinueBlockChain(nodeID)
			defer bc.Close()
			balance := 0
			UTXOSet := blockchain.UTXOSet{bc}
			pubKeyHash := util.Base58Decode([]byte(address))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-CheckSumLength]
			UTXOs := UTXOSet.FindUTXO(pubKeyHash)
			for _, out := range UTXOs {
				fmt.Printf("Out: %v\n", out)
				balance += out.Value
			}
			fmt.Printf("Balance of '%s': %d\n", address, balance)
		},
	}
	flags := cmd.PersistentFlags()
	flags.StringP("address", "a", "", "Address to inquirey balance")
	cmd.MarkPersistentFlagRequired("address")
	viper.BindPFlag("address", flags.Lookup("address"))
	return cmd
}
func PrintCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show",
		Short: "Prints blockchain",
		Run: func(_ *cobra.Command, _ []string) {
			nodeID := viper.GetString("NodeID")
			chain := blockchain.ContinueBlockChain(nodeID)
			chain.Print()
		},
	}
}
