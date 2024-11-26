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
			UTXO := blockchain.UTXOSet{BlockChain: bc}
			tx := ws.NewTransaction(from, to, amount, &UTXO)
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
