package node

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/storage"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/Amr-Shams/Blocker/wallet"
	"github.com/google/uuid"
	"github.com/spf13/viper"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	StateIdle = iota
	StateMining
	StateValidating
	StateConnected
)

var (
	availablePorts   = []string{"5001", "5002", "5003", "5004", "5005"}
	MAX_MEMPOOL_SIZE = 10
)

var stateName = map[int]string{
	StateIdle:      "Idle",
	StateMining:    "Mining",
	StateConnected: "Connected",
}

type WalletNode struct {
	ID             int32
	Address        string
	Status         int
	Blockchain     *blockchain.BlockChain
	WalletPeers    map[string]*WalletNode
	FullPeers      map[string]*FullNode
	Type           string
	Wallets        *wallet.Wallets
	ConnectionPool map[string]*grpc.ClientConn
	TSXhistory     []*TSXhistory
	pb.UnimplementedBlockchainServiceServer
}

var _ Node = &WalletNode{}
var dbMutex sync.Mutex

func (n *WalletNode) GetWalletPeers() map[string]*WalletNode {
	return n.WalletPeers
}
func (n *WalletNode) GetFullPeers() map[string]*FullNode {
	return n.FullPeers
}
func (n *WalletNode) CloseDB() {
	n.Blockchain.Close()
}
func (n *WalletNode) GetStatus() int {
	return n.Status
}
func (n *WalletNode) GetType() string {
	return n.Type
}
func (s *WalletNode) GetAllPeers() map[string]*WalletNode {
	peers := make(map[string]*WalletNode, len(s.FullPeers)+len(s.WalletPeers))
	for k, v := range s.FullPeers {
		peers[k] = &v.WalletNode
	}
	for k, v := range s.WalletPeers {
		peers[k] = v
	}
	return peers
}

func (n *WalletNode) GetWallets() *wallet.Wallets {
	return n.Wallets
}
func (n *WalletNode) connectToPeer(address string) (*grpc.ClientConn, error) {
	if n.ConnectionPool == nil {
		dbMutex.Lock()
		if n.ConnectionPool == nil {
			n.ConnectionPool = make(map[string]*grpc.ClientConn)
		}
		dbMutex.Unlock()
	}

	dbMutex.Lock()
	defer dbMutex.Unlock()

	if conn, exists := n.ConnectionPool[address]; exists {
		state := conn.GetState()
		if state == connectivity.Ready || state == connectivity.Idle {
			return conn, nil
		}
		conn.Close()
		delete(n.ConnectionPool, address)
	}

	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create connection to %s: %v", address, err)
	}

	n.ConnectionPool[address] = conn
	return conn, nil
}
func (n *WalletNode) DiscoverPeers() (bestPeer *FullNode) {
	n.ConnectionPool = make(map[string]*grpc.ClientConn)
	for _, port := range availablePorts {
		fullAddress := "localhost:" + port
		if fullAddress != n.Address {
			conn, err := n.connectToPeer(fullAddress)
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", fullAddress, err)
				continue
			}
			client := pb.NewBlockchainServiceClient(conn)
			peerResponse, err := client.Hello(context.Background(), &pb.Request{})
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", fullAddress, err)
				continue
			}

			peers, _ := client.GetPeers(context.Background(), &pb.Empty{})
			fmt.Printf("Peers Length in client: %d\n", len(peers.Peers))
			fmt.Println("Current node wallets length:", len(n.GetWallets().Wallets))
			_, err = client.AddPeer(context.Background(), &pb.AddPeerRequest{
				NodeId:         int32(n.ID),
				Address:        n.GetAddress(),
				Status:         stateName[n.Status],
				Type:           n.Type,
				Wallets:        mapBlockchainWalletsToServerWallets(n.GetWallets()),
				Height:         int64(n.Blockchain.GetHeight()),
				LastTimeUpdate: n.Blockchain.GetLastTimeUpdate(),
			})
			if err != nil {
				fmt.Printf("Failed to add peer %s: %v\n", fullAddress, err)
				continue
			}
			var WalletNode = &WalletNode{
				ID:             peerResponse.NodeId,
				Address:        peerResponse.Address,
				Status:         getStatusFromString(peerResponse.Status),
				Type:           peerResponse.Type,
				Wallets:        mapServerWalletsToBlockchainWallets(peerResponse.Wallets),
				Blockchain:     &blockchain.BlockChain{LastTimeUpdate: peerResponse.LastTimeUpdate, Height: peerResponse.Height},
				ConnectionPool: make(map[string]*grpc.ClientConn),
			}
			switch peerResponse.Type {
			case "full":
				n.FullPeers[fullAddress] = &FullNode{
					WalletNode: *WalletNode,
					MemPool:    blockchain.NewMemPool(MAX_MEMPOOL_SIZE),
				}
			case "wallet":
				n.WalletPeers[fullAddress] = WalletNode
			}
			currentPeer := n.FullPeers[fullAddress]
			if currentPeer == nil {
				continue
			}
			if bestPeer == nil {
				bestPeer = currentPeer
			} else if currentPeer.WalletNode.Blockchain != nil &&
				bestPeer.Blockchain != nil &&
				bestPeer.Blockchain.GetHeight() < currentPeer.WalletNode.Blockchain.GetHeight() {
				bestPeer = currentPeer
			}
		}
	}
	return bestPeer
}

func (n *WalletNode) FetchBlockchain(peer *FullNode) ([]*pb.Block, error) {
	if peer == nil || peer.GetAddress() == n.GetAddress() {
		return nil, nil
	}
	fmt.Printf("Fetching blockchain from peer %s\n", peer.GetAddress())
	conn, err := n.connectToPeer(peer.GetAddress())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %v", peer.GetAddress(), err)
	}
	client := pb.NewBlockchainServiceClient(conn)
	response, err := client.GetBlockChain(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blockchain from peer %s: %v", peer.GetAddress(), err)
	}

	fmt.Printf("Fetched blockchain from peer %s with %d blocks\n", peer.GetAddress(), len(response.Blocks))
	return response.Blocks, nil
}

func (n *WalletNode) UpdateBlockchain(newBlocks []*pb.Block) error {
	newBlocks = newBlocks[n.Blockchain.GetHeight():]
	UTXOSet := blockchain.UTXOSet{BlockChain: n.Blockchain}
	for _, protoBlock := range newBlocks {
		transactions := mapServerTSXsToBlockchainTSXs(protoBlock.Transactions)
		latestHash := n.Blockchain.GetLastHash()
		if !bytes.Equal(protoBlock.PrevHash, latestHash) {
			fmt.Println("PrevHash does not match the last block in the blockchain")
			continue
		}
		block := &blockchain.Block{
			TimeStamp:    time.Unix(protoBlock.Timestamp, 0),
			Hash:         protoBlock.Hash,
			Transactions: transactions,
			PrevHash:     protoBlock.PrevHash,
			Nonce:        int(protoBlock.Nonce),
		}

		proof := blockchain.NewProof(block)
		if !proof.Validate() {
			fmt.Println("Proof of work is invalid")
			continue
		}
		if err := n.Blockchain.AddEntireBlock(block); !err {
			fmt.Println("Failed to add block to blockchain")
			continue
		}
		UTXOSet.Update(block)
	}
	return nil
}

func (s *WalletNode) GetAddress() string {
	return s.Address
}

func (s *WalletNode) SetStatus(status int) {
	s.Status = status
}
func (s *WalletNode) GetBlockchain() *blockchain.BlockChain {
	return s.Blockchain
}
func (s *WalletNode) SetBlockchain(bc *blockchain.BlockChain) {
	fmt.Printf("Setting blockchain for %s\n", s.Address)
	s.Blockchain = bc
}
func (s *WalletNode) GetPort() string {
	return s.Address[strings.LastIndex(s.Address, ":")+1:]
}
func (s *WalletNode) AddWallet() {
	s.Wallets.AddWallet(s.GetAddress())
}

func (b *WalletNode) GetLastHash() []byte {
	if b.Blockchain == nil {
		return []byte{}
	}
	return b.Blockchain.GetLastHash()
}
func (b *WalletNode) AllPeersAddress() []string {
	peers := []string{}
	pairs := b.GetAllPeers()
	for _, k := range pairs {
		peers = append(peers, k.GetAddress())
	}
	return peers
}

func (b *WalletNode) GetTSXhistory() []*TSXhistory {
	return b.TSXhistory
}
func (b *WalletNode) SetTSXhistory(TSXhistory []*TSXhistory) {
	b.TSXhistory = TSXhistory
}

func NewWalletNode() *WalletNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	fs, err := storage.GetFileSystem()
	util.Handle(err)
	data, err := fs.ReadFromFile()
	util.Handle(err)
	TSXhistory := mapBytestoTSXhistory(data)
	return &WalletNode{
		ID:             int32(id),
		Address:        address,
		Blockchain:     blockchain.ContinueBlockChain(address),
		Status:         StateIdle,
		FullPeers:      make(map[string]*FullNode),
		WalletPeers:    make(map[string]*WalletNode),
		Type:           "wallet",
		Wallets:        wallets,
		ConnectionPool: make(map[string]*grpc.ClientConn),
		TSXhistory:     TSXhistory,
	}
}
