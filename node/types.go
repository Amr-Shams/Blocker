package node

import (
	"context"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/wallet"
	"google.golang.org/grpc"
)

type serverNode interface {
	GetBlockChain(context.Context, *pb.Empty) (*pb.Response, error)
	BroadcastTSX(context.Context, *pb.Transaction) (*pb.Response, error)
	AddPeer(context.Context, *pb.AddPeerRequest) (*pb.AddPeerResponse, error)
	GetPeers(context.Context, *pb.Empty) (*pb.GetPeersResponse, error)
	CheckStatus(context.Context, *pb.Request) (*pb.Response, error)
	DeleteTSXMempool(context.Context, *pb.DeleteTSXMempoolRequest) (*pb.Empty, error)
	MineBlock(context.Context, *pb.Empty) (*pb.Response, error)
	BroadcastBlock(context.Context, *pb.Block) (*pb.Response, error)
	AddTSXMempool(context.Context, *pb.Transaction) (*pb.Empty, error)
	AddBlock(context.Context, *pb.Block) (*pb.Response, error)
	Hello(context.Context, *pb.Request) (*pb.HelloResponse, error)
}
type Node interface {
	GetAddress() string
	SetStatus(int)
	GetStatus() int
	GetType() string
	GetBlockchain() *blockchain.BlockChain
	SetBlockchain(*blockchain.BlockChain)
	GetPort() string
	AllPeersAddress() []string
	DiscoverPeers() *FullNode
	FetchBlockchain(peer *FullNode) ([]*pb.Block, error)
	UpdateBlockchain([]*pb.Block) error
	GetWallets() *wallet.Wallets
	AddWallet()
	CloseDB()
	GetAllPeers() map[string]*WalletNode
	GetWalletPeers() map[string]*WalletNode
	GetFullPeers() map[string]*FullNode
	connectToPeer(string) (*grpc.ClientConn, error)
	SetTSXhistory([]*TSXhistory)
	GetTSXhistory() []*TSXhistory
}

type FUllNode interface {
	Node
	serverNode
	AddMemPool(*blockchain.Transaction) error
	GetMemPool() *blockchain.Transaction
	DeleteTSXMemPool()
}
type TSXhistory struct {
	From   string `json:"from"`
	To     string `json:"to"`
	Amount string `json:"amount"`
}
