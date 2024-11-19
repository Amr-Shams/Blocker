package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/Amr-Shams/Blocker/wallet"
	"github.com/google/uuid"
	"github.com/soheilhy/cmux"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/exp/rand"
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

// REFAC : Refactor main node.go into at Moudules
// Module for the node interface
// Module for the db connection(Huge)
// Module for the grpc server and client
// Module for the Nodes (Base, Wallet, Full)
// Module for the routers and handlers

// TEST(16): Add unit test for the code
// Learn how to create unit test for grpc server
// how to make an integration test for grpc server and rest api ?

//

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

type Node interface {
	GetBlockChain(context.Context, *pb.Empty) (*pb.Response, error)
	BroadcastTSX(context.Context, *pb.Transaction) (*pb.Response, error)
	AddPeer(context.Context, *pb.AddPeerRequest) (*pb.AddPeerResponse, error)
	GetPeers(context.Context, *pb.Empty) (*pb.GetPeersResponse, error)
	GetAddress() string
	CheckStatus(context.Context, *pb.Request) (*pb.Response, error)
	SetStatus(int)
	GetStatus() int
	GetType() string
	GetBlockchain() *blockchain.BlockChain
	SetBlockchain(*blockchain.BlockChain)
	GetPort() string
	AddBlock(context.Context, *pb.Block) (*pb.Response, error)
	Hello(context.Context, *pb.Request) (*pb.HelloResponse, error)
	AllPeersAddress() []string
	DiscoverPeers() *FullNode
	FetchBlockchain(peer *FullNode) ([]*pb.Block, error)
	UpdateBlockchain([]*pb.Block) error
	GetWallets() *wallet.Wallets
	AddWallet(*wallet.Wallet)
	BroadcastBlock(context.Context, *pb.Block) (*pb.Response, error)
	AddTSXMempool(context.Context, *pb.Transaction) (*pb.Empty, error)
	CloseDB()
	GetAllPeers() map[string]*BaseNode
	GetWalletPeers() map[string]*WalletNode
	GetFullPeers() map[string]*FullNode
	MineBlock(context.Context, *pb.Empty) (*pb.Response, error)
	connectToPeer(string) (*grpc.ClientConn, error)
	DeleteTSXMempool(context.Context, *pb.DeleteTSXMempoolRequest) (*pb.Empty, error)
}
type FUllNode interface {
	Node
	AddMemPool(*blockchain.Transaction) error
	GetMemPool() *blockchain.Transaction
	DeleteTSXMemPool()
}

type BaseNode struct {
	ID             int32
	Address        string
	Status         int
	Blockchain     *blockchain.BlockChain
	WalletPeers    map[string]*WalletNode
	FullPeers      map[string]*FullNode
	Type           string
	Wallets        *wallet.Wallets
	ConnectionPool map[string]*grpc.ClientConn
	pb.UnimplementedBlockchainServiceServer
}

type FullNode struct {
	BaseNode
	MemPool *blockchain.MemPool
}

type WalletNode struct {
	BaseNode
}

var dbMutex sync.Mutex

func (n *BaseNode) GetWalletPeers() map[string]*WalletNode {
	return n.WalletPeers
}
func (n *BaseNode) GetFullPeers() map[string]*FullNode {
	return n.FullPeers
}
func (n *BaseNode) CloseDB() {
	n.Blockchain.Close()
}
func (n *BaseNode) GetStatus() int {
	return n.Status
}
func (n *BaseNode) GetType() string {
	return n.Type
}

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
func (s *BaseNode) GetBlockChain(ctx context.Context, re *pb.Empty) (*pb.Response, error) {
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
		dbMutex.Unlock()
	}()
	dbMutex.Lock()
	response := &pb.Response{
		Blocks:  mapClientBlocksToServerBlocks(s.Blockchain.GetBlocks()),
		Status:  stateName[s.Status],
		Success: true,
	}

	return response, nil
}
func (s *FullNode) DeleteTSXMempool(ctx context.Context, in *pb.DeleteTSXMempoolRequest) (*pb.Empty, error) {
	s.DeleteTSXMemPool(in.Id)
	return &pb.Empty{}, nil
}
func (s *FullNode) MineBlock(ctx context.Context, in *pb.Empty) (*pb.Response, error) {
	s.Status = StateMining
	defer func() {
		s.Status = StateIdle
	}()
	fmt.Println("Mining Block from mempool")
	tsx := s.GetMemPool()
	if tsx == nil {
		return &pb.Response{Success: false}, nil
	}
	UTXOSet := blockchain.UTXOSet{BlockChain: s.Blockchain}
	block, err := s.Blockchain.AddBlock([]*blockchain.Transaction{tsx})
	if !err {
		return &pb.Response{Success: false}, nil
	}
	addedBlock := s.Blockchain.GetBlock(block.Hash)
	addedBlock.Print()
	fmt.Println("Reindexing UTXO Set")
	UTXOSet.Update(block)
	fmt.Println("Deleting transaction from mempool")
	s.DeleteTSXMemPool(tsx.ID)
	serverBlocks := mapClientBlocksToServerBlocks([]*blockchain.Block{block})
	return &pb.Response{Success: true, Blocks: serverBlocks}, nil
}
func (s *FullNode) AddTSXMempool(ctx context.Context, in *pb.Transaction) (*pb.Empty, error) {
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
	}()
	dbMutex.Lock()
	defer dbMutex.Unlock()
	tsx := mapServerTsxToBlockchainTsx(in)
	err := s.AddMemPool(tsx)
	if err != nil {
		return &pb.Empty{}, err
	}
	fmt.Println("Transaction added to mempool")
	return &pb.Empty{}, nil
}
func MergeBlockChain(blockchain1, blockchain2 []*pb.Block) []*pb.Block {
	blockMap := make(map[string]*pb.Block)

	addToMap := func(blocks []*pb.Block) {
		for _, block := range blocks {
			blockHash := string(block.Hash)
			if existing, exists := blockMap[blockHash]; !exists ||
				(exists && shouldPreferBlock(block, existing)) {
				blockMap[blockHash] = block
			}
		}
	}

	addToMap(blockchain1)
	addToMap(blockchain2)
	var latestBlock *pb.Block
	for _, block := range blockMap {
		if latestBlock == nil || shouldPreferBlock(block, latestBlock) {
			latestBlock = block
		}
	}

	var mergedChain []*pb.Block
	currentBlock := latestBlock
	for currentBlock != nil && len(currentBlock.PrevHash) > 0 {
		fmt.Print("Current Block: %x\n", currentBlock.Hash)
		mergedChain = append([]*pb.Block{currentBlock}, mergedChain...)
		currentBlock = blockMap[string(currentBlock.PrevHash)]
	}
	if currentBlock != nil && len(currentBlock.PrevHash) == 0 {
		mergedChain = append([]*pb.Block{currentBlock}, mergedChain...)
	}
	return mergedChain
}

func shouldPreferBlock(block1, block2 *pb.Block) bool {
	if block1.Timestamp != block2.Timestamp {
		return block1.Timestamp > block2.Timestamp
	}

	if block1.Nonce != block2.Nonce {
		return block1.Nonce > block2.Nonce
	}

	return string(block1.Hash) > string(block2.Hash)
}

func (s *BaseNode) BroadcastBlock(ctx context.Context, block *pb.Block) (*pb.Response, error) {
	s.SetStatus(StateConnected)
	defer func() {
		s.SetStatus(StateIdle)
	}()
	var wg sync.WaitGroup
	for _, peer := range s.FullPeers {
		wg.Add(1)
		go func(peer *FullNode) {
			defer wg.Done()
			conn, err := s.connectToPeer(peer.GetAddress())
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", peer.GetAddress(), err)
				return
			}
			client := pb.NewBlockchainServiceClient(conn)
			added, err := client.AddBlock(ctx, block)
			if added.Success {
				fmt.Printf("Block sent to %s\n", peer.GetAddress())
				peer.BroadcastBlock(ctx, block)
			} else {
				fmt.Printf("Block not sent to %s\n", peer.GetAddress())
			}
			for _, tsx := range block.Transactions {
				client.DeleteTSXMempool(ctx, &pb.DeleteTSXMempoolRequest{Id: tsx.Id})
			}
		}(peer)
	}
	for _, peer := range s.WalletPeers {
		wg.Add(1)
		go func(peer *WalletNode) {
			defer wg.Done()
			fmt.Printf("Sending block to %s\n", peer.GetAddress())
			conn, err := s.connectToPeer(peer.GetAddress())
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", peer.GetAddress(), err)
				return
			}
			client := pb.NewBlockchainServiceClient(conn)
			added, err := client.AddBlock(ctx, block)
			if added.Success {
				fmt.Printf("Block sent to %s\n", peer.GetAddress())
			} else {
				fmt.Printf("Block not sent to %s\n", peer.GetAddress())
			}
		}(peer)
	}
	wg.Wait()
	return &pb.Response{Status: stateName[s.Status], Success: true}, nil
}

// @input: ctx context.Context - context for the operation
// @input: in *pb.Transaction - transaction to broadcast
// @return: chan *pb.Response - channel for response
// @return: chan error - channel for errors
func (s *BaseNode) BroadcastTSX(ctx context.Context, in *pb.Transaction) (*pb.Response, error) {
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
	}()
	var wg sync.WaitGroup
	errChan := make(chan error, len(s.FullPeers))
	for _, peer := range s.FullPeers {
		wg.Add(1)
		go func(peer *FullNode) {
			defer wg.Done()
			if err := s.broadcastToFullNode(ctx, in, peer); err != nil {
				errChan <- err
			} else {
				peer.BroadcastTSX(ctx, in)
			}
		}(peer)
	}
	wg.Wait()
	close(errChan)
	if len(errChan) > 0 {
		return &pb.Response{Status: stateName[s.Status], Success: false}, <-errChan
	}

	return &pb.Response{Status: stateName[s.Status], Success: true}, nil
}

func (s *BaseNode) broadcastToFullNode(ctx context.Context, in *pb.Transaction, peer *FullNode) error {
	fmt.Printf("Sending TSX to full node %s\n", peer.GetAddress())
	conn, err := s.connectToPeer(peer.GetAddress())
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %v", peer.GetAddress(), err)
	}
	client := pb.NewBlockchainServiceClient(conn)
	_, err = client.AddTSXMempool(ctx, in)
	if err != nil {
		return fmt.Errorf("failed to AddTSXMempool for %s: %v", peer.GetAddress(), err)
	}
	return nil
}

// TODO(17): Ignore the WalletNode for now
// it is just used to notify the wallet about the transaction

// func (s *BaseNode) broadcastToWalletNode(ctx context.Context, in *pb.Transaction, peer *WalletNode) error {
//     fmt.Printf("Sending TSX to wallet node %s\n", peer.GetAddress())
//     conn, err := s.connectToPeer(peer.GetAddress())
//     if err != nil {
//         return fmt.Errorf("failed to connect to %s: %v", peer.GetAddress(), err)
//     }
//     client := pb.NewBlockchainServiceClient(conn)

//     // For wallet nodes, we only notify about the transaction
//     _, err = client.NotifyTransaction(ctx, in) // You'll need to add this method to your protobuf service
//     if err != nil {
//         return fmt.Errorf("failed to notify wallet %s: %v", peer.GetAddress(), err)
//     }

//	    return nil
//	}
func (s *BaseNode) GetAddress() string {
	return s.Address
}
func (s *BaseNode) CheckStatus(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	return &pb.Response{Status: stateName[s.Status], Success: true}, nil
}
func (s *BaseNode) SetStatus(status int) {
	s.Status = status
}
func (s *BaseNode) GetBlockchain() *blockchain.BlockChain {
	return s.Blockchain
}
func (s *BaseNode) SetBlockchain(bc *blockchain.BlockChain) {
	fmt.Printf("Setting blockchain for %s\n", s.Address)
	s.Blockchain = bc
}
func (s *BaseNode) GetPort() string {
	return s.Address[strings.LastIndex(s.Address, ":")+1:]
}
func (s *BaseNode) AddWallet(w *wallet.Wallet) {
	s.Wallets.Wallets[string(w.PublicKey)] = w
}
func (s *BaseNode) AddPeer(ctx context.Context, in *pb.AddPeerRequest) (*pb.AddPeerResponse, error) {
	baseNode := BaseNode{
		ID:      in.NodeId,
		Address: in.Address,
		Status:  getStatusFromString(in.Status),
		Type:    in.Type,
		Wallets: mapServerWalletsToBlockchainWallets(in.Wallets),
		Blockchain: &blockchain.BlockChain{
			LastTimeUpdate: in.LastTimeUpdate,
			Height:         int(in.Height),
		},
	}
	switch in.Type {
	case "full":
		peer := &FullNode{BaseNode: baseNode}
		s.FullPeers[in.Address] = peer
		if s.Blockchain.Height < int(in.Height) {
			go func() {
				syncCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				defer cancel()
				syncDone := make(chan error, 1)
				go func() {
					err := SyncBlockchain(s, peer)
					syncDone <- err
				}()
				select {
				case err := <-syncDone:
					if err != nil {
						log.Printf("Failed to sync blockchain with %s: %v", in.Address, err)
					}
				case <-syncCtx.Done():
					log.Printf("Sync operation with %s timed out", in.Address)
				}
			}()
		}
	case "wallet":
		s.WalletPeers[in.Address] = &WalletNode{BaseNode: baseNode}
	default:
		return nil, fmt.Errorf("unknown peer type: %s", in.Type)
	}
	return &pb.AddPeerResponse{Success: true}, nil
}
func (s *BaseNode) AddBlock(ctx context.Context, in *pb.Block) (*pb.Response, error) {
	s.SetStatus(StateConnected)
	defer func() {
		s.SetStatus(StateIdle)
	}()
	fmt.Println("Received new block")
	dbMutex.Lock()
	defer dbMutex.Unlock()
	block := &blockchain.Block{
		TimeStamp:    time.Unix(in.Timestamp, 0),
		Hash:         in.Hash,
		Transactions: mapServerTSXsToBlockchainTSXs(in.Transactions),
		PrevHash:     in.PrevHash,
		Nonce:        int(in.Nonce),
	}
	proof := blockchain.NewProof(block)
	if !proof.Validate() {
		return &pb.Response{Success: false}, nil
	}
	if err := s.Blockchain.AddEntireBlock(block); !err {
		return &pb.Response{Success: false}, nil
	}
	UTXOSet := blockchain.UTXOSet{BlockChain: s.GetBlockchain()}
	UTXOSet.Update(block)
	return &pb.Response{Success: true}, nil
}
func mapServerWalletToBlockchainWallet(w *pb.Wallet) *wallet.Wallet {
	return &wallet.Wallet{
		PublicKey: w.PublicKey,
	}
}
func mapBlockchainWalletToServerWallet(wallet *wallet.Wallet) *pb.Wallet {
	if wallet == nil {
		return &pb.Wallet{}
	}
	return &pb.Wallet{
		PublicKey: wallet.PublicKey,
	}
}
func mapServerWalletsToBlockchainWallets(wallets *pb.Wallets) *wallet.Wallets {
	ws := &wallet.Wallets{Wallets: make(map[string]*wallet.Wallet)}
	for _, w := range wallets.Wallets {
		ws.Wallets[string(w.PublicKey)] = mapServerWalletToBlockchainWallet(w)
	}
	return ws
}
func mapBlockchainWalletsToServerWallets(wallets *wallet.Wallets) *pb.Wallets {
	ws := &pb.Wallets{}
	if wallets == nil || wallets.Wallets == nil {
		return ws
	}
	for _, w := range wallets.Wallets {
		ws.Wallets = append(ws.Wallets, mapBlockchainWalletToServerWallet(w))
	}
	return ws
}
func (s *BaseNode) Hello(ctx context.Context, in *pb.Request) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		NodeId: s.ID, Address: s.Address, Status: stateName[s.Status], Type: s.Type, LastHash: s.Blockchain.LastHash, LastTimeUpdate: s.Blockchain.GetLastTimeUpdate(),
		Wallets: mapBlockchainWalletsToServerWallets(s.GetWallets()), Height: int64(s.Blockchain.GetHeight()),
	}, nil
}

func getStatusFromString(status string) int {
	switch status {
	case "Idle":
		return StateIdle
	case "Mining":
		return StateMining
	case "Connected":
		return StateConnected
	default:
		return StateIdle
	}
}
func (s *BaseNode) GetPeers(ctx context.Context, in *pb.Empty) (*pb.GetPeersResponse, error) {
	response := []*pb.HelloResponse{}
	for _, peer := range s.FullPeers {
		height := int64(0)
		lastTimeUpdate := int64(0)
		if peer.Blockchain != nil {
			height = int64(peer.Blockchain.GetHeight())
			lastTimeUpdate = peer.Blockchain.GetLastTimeUpdate()
		}
		response = append(response, &pb.HelloResponse{
			NodeId:         peer.ID,
			Address:        peer.Address,
			Status:         stateName[peer.Status],
			Type:           peer.Type,
			LastHash:       peer.GetLastHash(),
			Height:         height,
			LastTimeUpdate: lastTimeUpdate,
			Wallets:        mapBlockchainWalletsToServerWallets(peer.GetWallets()),
		})
	}
	for _, peer := range s.WalletPeers {
		height := int64(0)
		lastTimeUpdate := int64(0)
		if peer.Blockchain != nil {
			height = int64(peer.Blockchain.GetHeight())
			lastTimeUpdate = peer.Blockchain.GetLastTimeUpdate()
		}
		response = append(response, &pb.HelloResponse{
			NodeId:         peer.ID,
			Address:        peer.Address,
			Status:         stateName[peer.Status],
			Type:           peer.Type,
			LastHash:       peer.GetLastHash(),
			Height:         height,
			LastTimeUpdate: lastTimeUpdate,
			Wallets:        mapBlockchainWalletsToServerWallets(peer.GetWallets()),
		})
	}
	return &pb.GetPeersResponse{Peers: response}, nil
}
func (b *BaseNode) GetLastHash() []byte {
	if b.Blockchain == nil {
		return []byte{}
	}
	return b.Blockchain.GetLastHash()
}
func (b *BaseNode) AllPeersAddress() []string {
	peers := []string{}
	pairs := b.GetAllPeers()
	for _, k := range pairs {
		peers = append(peers, k.GetAddress())
	}
	return peers
}

func NewFullNode() *FullNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	return &FullNode{
		BaseNode: BaseNode{
			ID:             int32(id),
			Address:        address,
			Blockchain:     blockchain.ContinueBlockChain(address),
			Status:         StateIdle,
			FullPeers:      make(map[string]*FullNode),
			WalletPeers:    make(map[string]*WalletNode),
			Type:           "full",
			Wallets:        wallets,
			ConnectionPool: make(map[string]*grpc.ClientConn),
		},
		MemPool: blockchain.NewMemPool(MAX_MEMPOOL_SIZE),
	}
}

func NewWalletNode() *WalletNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	return &WalletNode{
		BaseNode: BaseNode{
			ID:             int32(id),
			Address:        address,
			Blockchain:     blockchain.ContinueBlockChain(address),
			Status:         StateIdle,
			FullPeers:      make(map[string]*FullNode),
			WalletPeers:    make(map[string]*WalletNode),
			Type:           "wallet",
			Wallets:        wallets,
			ConnectionPool: make(map[string]*grpc.ClientConn),
		},
	}
}
func (n *BaseNode) connectToPeer(address string) (*grpc.ClientConn, error) {
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
func (n *BaseNode) DiscoverPeers() (bestPeer *FullNode) {

	// Step 1: Gather metadata from each peer
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

			// Send metadata about the current node to the peer
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
			switch peerResponse.Type {
			case "full":
				n.FullPeers[fullAddress] = &FullNode{
					BaseNode: BaseNode{
						ID:      peerResponse.NodeId,
						Address: peerResponse.Address,
						Status:  getStatusFromString(peerResponse.Status),
						Type:    peerResponse.Type,
						Wallets: mapServerWalletsToBlockchainWallets(peerResponse.Wallets),
						Blockchain: &blockchain.BlockChain{
							LastTimeUpdate: peerResponse.LastTimeUpdate,
							Height:         int(peerResponse.Height),
							LastHash:       peerResponse.LastHash,
						},
					},
				}
			case "wallet":
				n.WalletPeers[fullAddress] = &WalletNode{
					BaseNode: BaseNode{
						ID:      peerResponse.NodeId,
						Address: peerResponse.Address,
						Status:  getStatusFromString(peerResponse.Status),
						Type:    peerResponse.Type,
						Wallets: mapServerWalletsToBlockchainWallets(peerResponse.Wallets),
						Blockchain: &blockchain.BlockChain{
							LastTimeUpdate: peerResponse.LastTimeUpdate,
							Height:         int(peerResponse.Height),
							LastHash:       peerResponse.LastHash,
						},
					},
				}
			}
			// Identify the "best" peer based on blockchain height
			currentPeer := n.FullPeers[fullAddress]
			if currentPeer == nil {
				continue // Skip nil peers
			}

			if bestPeer == nil {
				bestPeer = currentPeer
			} else if currentPeer.BaseNode.Blockchain != nil &&
				bestPeer.Blockchain != nil &&
				bestPeer.Blockchain.GetHeight() < currentPeer.BaseNode.Blockchain.GetHeight() {
				bestPeer = currentPeer
			}
		}
	}
	return bestPeer
}

// @input: peer *BaseNode - peer node to fetch blockchain from
// @return: []*pb.Block - array of blocks fetched from peer
// @return: error - error if any occurred during fetch
func (n *BaseNode) FetchBlockchain(peer *FullNode) ([]*pb.Block, error) {
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

// @input: newBlocks []*pb.Block - array of new blocks to be added to blockchain
// @return: error - error if any occurred during update
func (n *BaseNode) UpdateBlockchain(newBlocks []*pb.Block) error {

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

// @input: n Node - node to sync blockchain for
// @input: bestPeer *BaseNode - peer to sync blockchain from
// @return: error - error if any occurred during sync
func SyncBlockchain(n Node, bestPeer *FullNode) error {
	defer func() {
		// n.CloseDB()
	}()
	blocks, err := n.FetchBlockchain(bestPeer)
	if blocks == nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to fetch blockchain: %v", err)
	}
	if len(blocks) < n.GetBlockchain().GetHeight() {
		return nil
	}
	if err := n.UpdateBlockchain(blocks); err != nil {
		return fmt.Errorf("failed to update blockchain: %v", err)
	}

	fmt.Printf("Chain of Node %s has %d blocks\n",
		n.GetAddress(), len(n.GetBlockchain().GetBlocks()))

	return nil
}
func grpcServer(l net.Listener, n Node) error {
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	pb.RegisterBlockchainServiceServer(grpcServer, n)
	return grpcServer.Serve(l)
}
func (s *BaseNode) GetAllPeers() map[string]*BaseNode {
	peers := make(map[string]*BaseNode, len(s.FullPeers)+len(s.WalletPeers))
	for k, v := range s.FullPeers {
		peers[k] = &v.BaseNode
	}
	for k, v := range s.WalletPeers {
		peers[k] = &v.BaseNode
	}
	return peers
}
func mapServerTsxToBlockchainTsx(tsx *pb.Transaction) *blockchain.Transaction {
	Vin := []blockchain.TXInput{}
	for _, input := range tsx.Vin {
		Vin = append(Vin, blockchain.TXInput{
			Txid:      input.Txid,
			Vout:      int(input.Vout),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}
	Vout := []blockchain.TXOutput{}
	for _, output := range tsx.Vout {
		Vout = append(Vout, blockchain.TXOutput{
			Value:      int(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}
	return &blockchain.Transaction{
		ID:   tsx.Id,
		Vin:  Vin,
		Vout: Vout,
	}
}
func mapBlockchainTsxToServerTsx(tsx *blockchain.Transaction) *pb.Transaction {
	Vin := []*pb.TXInput{}
	for _, input := range tsx.Vin {
		Vin = append(Vin, &pb.TXInput{
			Txid:      input.Txid,
			Vout:      int32(input.Vout),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}
	Vout := []*pb.TXOutput{}
	for _, output := range tsx.Vout {
		Vout = append(Vout, &pb.TXOutput{
			Value:      int32(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}
	return &pb.Transaction{
		Id:   tsx.ID,
		Vin:  Vin,
		Vout: Vout,
	}
}
func mapClientBlocksToServerBlocks(blocks []*blockchain.Block) []*pb.Block {
	pbBlocks := []*pb.Block{}
	for _, block := range blocks {
		pbBlocks = append(pbBlocks, &pb.Block{
			Hash:         block.Hash,
			PrevHash:     block.PrevHash,
			Transactions: mapBlockchainTSXsToServerTSXs(block.Transactions),
			Nonce:        int32(block.Nonce),
			Timestamp:    block.TimeStamp.Unix(),
		})
	}
	return pbBlocks
}
func mapServerBlocksToClientBlocks(blocks []*pb.Block) []*blockchain.Block {
	clientBlocks := []*blockchain.Block{}
	for _, block := range blocks {
		clientBlocks = append(clientBlocks, &blockchain.Block{
			Hash:         block.Hash,
			PrevHash:     block.PrevHash,
			Transactions: mapServerTSXsToBlockchainTSXs(block.Transactions),
			Nonce:        int(block.Nonce),
			TimeStamp:    time.Unix(block.Timestamp, 0),
		})
	}
	return clientBlocks
}
func mapServerTSXsToBlockchainTSXs(tsxs []*pb.Transaction) []*blockchain.Transaction {
	res := []*blockchain.Transaction{}
	for _, tsx := range tsxs {
		res = append(res, mapServerTsxToBlockchainTsx(tsx))
	}
	return res
}
func mapBlockchainTSXsToServerTSXs(tsxs []*blockchain.Transaction) []*pb.Transaction {
	pbTSXs := []*pb.Transaction{}
	for _, tsx := range tsxs {
		pbTSXs = append(pbTSXs, mapBlockchainTsxToServerTsx(tsx))
	}
	return pbTSXs
}
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
func (n *BaseNode) GetWallets() *wallet.Wallets {
	return n.Wallets
}

func httpServer(l net.Listener, n Node) error {
	mux := http.NewServeMux()
	// create a get request to get all paris
	// create a get request to get info about the node

	mux.HandleFunc("/pairs", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// n.CloseDB()
		}()
		pairs := n.GetAllPeers()
		response, _ := json.Marshal(pairs)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// n.CloseDB()
		}()
		response, _ := json.Marshal(n)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// n.CloseDB()
		}()
		response, _ := json.Marshal(stateName[n.GetStatus()])
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		ws := n.GetWallets()
		var balances = make(map[string]int)
		UTXOSet := blockchain.UTXOSet{BlockChain: n.GetBlockchain()}
		for _, w := range ws.Wallets {

			pubKeyHash := util.Base58Decode([]byte(w.GetAddress()))
			pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.CheckSumLength]
			UTXOs := UTXOSet.FindUTXO(pubKeyHash)
			balance := 0
			for _, out := range UTXOs {
				fmt.Printf("Out: %v\n", out)
				balance += out.Value
			}
			fmt.Printf("Address: %s, Balance: %d\n", w.GetAddress(), balance)
			balances[w.GetAddress()] = balance
		}
		response, _ := json.Marshal(balances)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		fromWallet := r.FormValue("from")
		toWallet := r.FormValue("to")
		amountStr := r.FormValue("amount")
		amount := util.StrToInt(amountStr)
		if !util.ValidateAddress(fromWallet) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid from address"))
			return
		}
		if !util.ValidateAddress(toWallet) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid to address"))
			return
		}
		if !util.ValidateAmount(amount) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid amount"))
			return
		}
		ws := n.GetWallets()
		if ws == nil || ws.Authenticated(fromWallet) == false {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte("Unauthorized"))
			return
		}
		fmt.Println("Creating transaction")
		UTXO := blockchain.UTXOSet{
			BlockChain: n.GetBlockchain(),
		}
		tx := ws.NewTransaction(fromWallet, toWallet, amount, &UTXO)
		if tx == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Transaction failed"))
		}
		if fullNode, ok := n.(*FullNode); ok {
			fmt.Println("Adding transaction to mempool")
			err := fullNode.AddMemPool(tx)
			fmt.Println(err)
			if err == nil {
				fmt.Println("Broadcasting transaction")
				done := make(chan bool) // we communicate by sharing memory
				go func() {
					fullNode.BroadcastTSX(context.Background(), mapBlockchainTsxToServerTsx(tx))
					done <- true
				}()
				<-done
			} else {
				util.Handle(err)
			}
		} else {
			done := make(chan bool)
			go func() {
				n.BroadcastTSX(context.Background(), mapBlockchainTsxToServerTsx(tx))
				done <- true
			}()
			<-done
		}
		miningNode := SelectRandomFullNode(n)
		if miningNode != nil {
			conn, err := n.connectToPeer(miningNode.GetAddress())
			util.Handle(err)
			client := pb.NewBlockchainServiceClient(conn)
			res, err := client.MineBlock(context.Background(), &pb.Empty{})
			util.Handle(err)
			client.BroadcastBlock(context.Background(), res.Blocks[0])
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("No mining node available"))
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction sent"))
	})
	mux.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			// n.CloseDB()
		}()
		ws := n.GetWallets()
		response, _ := json.Marshal(ws)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/AddWallet", func(w http.ResponseWriter, r *http.Request) {
		wallet := wallet.NewWallet()
		n.AddWallet(wallet)
		response, _ := json.Marshal(wallet)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/memPool", func(w http.ResponseWriter, r *http.Request) {
		if fullNode, ok := n.(*FullNode); ok {
			memPool := fullNode.GetMemPool()
			response, _ := json.Marshal(memPool)
			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
		} else {
			w.WriteHeader(http.StatusMethodNotAllowed)
			w.Write([]byte("Method not allowed"))
		}
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		cwd, err := os.Getwd()
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		http.ServeFile(w, r, filepath.Join(cwd, "views", "home.html"))
	})
	s := &http.Server{
		Handler: enableCORS(mux),
	}
	return s.Serve(l)
}

func SelectRandomFullNode(n Node) *FullNode {
	peers := n.GetFullPeers()
	if len(peers) == 0 {
		return nil
	}
	idx := rand.Intn(len(peers))
	for _, peer := range peers {
		if idx == 0 {
			if peer.GetStatus() == StateIdle {
				return peer
			} else {
				idx = rand.Intn(len(peers))
			}
		}
		idx--
	}
	return nil
}

func StartWalletNodeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "run2",
		Short:   "Start a wallet node",
		Long:    "a wallet node that can send and receive transactions",
		Example: "wallet print",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			n := NewWalletNode()
			bestPeer := n.DiscoverPeers()
			fmt.Printf("Node is connected to %d peers\n", len(n.AllPeersAddress()))
			if bestPeer != nil {
				fmt.Printf("Best peer is %s\n", bestPeer.GetAddress())
				fmt.Printf("Height is %d\n", bestPeer.Blockchain.GetHeight())
			}
			go func() {
				if err := SyncBlockchain(&n.BaseNode, bestPeer); err != nil {
					log.Printf("Error syncing blockchain: %v", err)
				}
			}()

			// Set up server
			port := n.GetPort()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				log.Fatalf("Failed to listen on port %s: %v", port, err)
			}

			conn, err := grpc.NewClient(n.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to create gRPC connection: %v", err)
			}
			defer conn.Close()

			n.SetStatus(StateConnected)
			defer func() {
				n.SetStatus(StateIdle)
			}()

			// Initialize cmux
			m := cmux.New(lis)

			// Use specific matchers for gRPC and HTTP
			grpcL := m.MatchWithWriters(
				cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			)
			httpL := m.Match(cmux.HTTP1Fast())

			// Create error group for managing goroutines
			g := new(errgroup.Group)

			// Start gRPC server
			g.Go(func() error {
				log.Printf("Starting gRPC server on port %s\n", port)
				if err := grpcServer(grpcL, n); err != nil {
					return fmt.Errorf("gRPC server error: %v", err)
				}
				return nil
			})

			// Start HTTP server
			g.Go(func() error {
				log.Printf("Starting HTTP server on port %s\n", port)
				if err := httpServer(httpL, n); err != nil {
					return fmt.Errorf("HTTP server error: %v", err)
				}
				return nil
			})

			// Start cmux
			g.Go(func() error {
				log.Printf("Starting multiplexer on port %s\n", port)
				return m.Serve()
			})

			// Wait for all servers and handle errors
			if err := g.Wait(); err != nil {
				log.Fatalf("Server error: %v", err)
			}
		},
	}
}

func StartFullNodeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "run1",
		Short:   "Start a full node",
		Long:    "A full node that can mine blocks and validate transactions",
		Example: "node print",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			n := NewFullNode()
			b := n.DiscoverPeers()
			fmt.Printf("Node is connected to %d peers\n", len(n.AllPeersAddress()))

			conn, err := grpc.NewClient(n.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			util.Handle(err)
			defer conn.Close()

			n.SetStatus(StateConnected)
			defer func() {
				n.SetStatus(StateIdle)
			}()
			go func() {
				if err := SyncBlockchain(n, b); err != nil {
					log.Printf("Error syncing blockchain: %v", err)
				}
			}()
			port := n.GetPort()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				log.Fatalf("Failed to listen on port %s: %v", port, err)
			}

			m := cmux.New(lis)
			grpcL := m.MatchWithWriters(
				cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"),
			)
			httpL := m.Match(cmux.HTTP1Fast())

			g := new(errgroup.Group)

			g.Go(func() error {
				log.Printf("Starting gRPC server on port %s\n", port)
				if err := grpcServer(grpcL, n); err != nil {
					return fmt.Errorf("gRPC server error: %v", err)
				}
				return nil
			})

			g.Go(func() error {
				log.Printf("Starting HTTP server on port %s\n", port)
				if err := httpServer(httpL, n); err != nil {
					return fmt.Errorf("HTTP server error: %v", err)
				}
				return nil
			})

			g.Go(func() error {
				log.Printf("Starting multiplexer on port %s\n", port)
				return m.Serve()
			})

			if err := g.Wait(); err != nil {
				log.Fatalf("Server error: %v", err)
			}
		},
	}
}
