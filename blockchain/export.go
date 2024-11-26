package blockchain

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
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
		Run: func(cmd *cobra.Command, args []string) {
			nodeID := viper.GetString("NodeID")
			chain := ContinueBlockChain(nodeID)
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			chain.Print()
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			cmd.Println(buf.String())
		},
	}
}

func CreateBlockChainCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new blockchain",
		Run: func(cmd *cobra.Command, _ []string) {
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
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			fmt.Println("Blockchain created")
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			cmd.Println(buf.String())
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
		Run: func(cmd *cobra.Command, _ []string) {
			nodeID := viper.GetString("NodeID")
			db, _ := storage.GetInstance(nodeID)
			defer db.Clean()
			oldStdout := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w
			fmt.Println("Blockchain cleaned")
			w.Close()
			os.Stdout = oldStdout
			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			cmd.Println(buf.String())
		},
	}
}
