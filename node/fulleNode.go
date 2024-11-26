package node

import (
	"github.com/Amr-Shams/Blocker/blockchain"
	"github.com/Amr-Shams/Blocker/storage"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/Amr-Shams/Blocker/wallet"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
)

type FullNode struct {
	WalletNode
	MemPool *blockchain.MemPool
}

var _ Node = &FullNode{}
var _ serverNode = &FullNode{}

func (n *FullNode) AddMemPool(tsx *blockchain.Transaction) error {
	n.Status = StateMining
	defer func() {
		n.Status = StateIdle
	}()
	return n.MemPool.AddTx(tsx)
}
func (n *FullNode) GetMemPool() *blockchain.Transaction {
	return n.MemPool.GetTx()
}

func (n *FullNode) DeleteTSXMemPool(ID []byte) {
	n.MemPool.DeleteTx(ID)
}

func NewFullNode() *FullNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	fs, err := storage.GetFileSystem()
	util.Handle(err)
	data, err := fs.ReadFromFile()
	util.Handle(err)
	TSXhistory := mapBytestoTSXhistory(data)
	return &FullNode{
		WalletNode: WalletNode{
			ID:             int32(id),
			Address:        address,
			Blockchain:     blockchain.ContinueBlockChain(address),
			Status:         StateIdle,
			FullPeers:      make(map[string]*FullNode),
			WalletPeers:    make(map[string]*WalletNode),
			Type:           "full",
			Wallets:        wallets,
			ConnectionPool: make(map[string]*grpc.ClientConn),
			TSXhistory:     TSXhistory,
		},
		MemPool: blockchain.NewMemPool(MAX_MEMPOOL_SIZE),
	}
}
