package node

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
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
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
)

//
//

// FIXMEEEE(11): BroadcastBlock/TSX are not working
// possible solution: check the logic of floating the transactions to the network
const (
	StateIdle = iota
	StateMining
	StateValidating
	StateConnected
)

var (
	availablePorts = []string{"5001", "5002", "5003", "5004", "5005"}
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
	DiscoverPeers() *BaseNode
	FetchBlockchain(peer *BaseNode) ([]*pb.Block, error)
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
}
type FUllNode interface {
	Node
	AddMemPool(*blockchain.Transaction) error
	GetMemPool() *blockchain.Transaction
	DeleteTSXMemPool()
}

type BaseNode struct {
	ID          int32
	Address     string
	Status      int
	Blockchain  *blockchain.BlockChain
	WalletPeers map[string]*WalletNode
	FullPeers   map[string]*FullNode
	Type        string
	Wallets     *wallet.Wallets
	pb.UnimplementedBlockchainServiceServer
}

type FullNode struct {
	BaseNode
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
	n.Blockchain.Database.Close()
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
		n.CloseDB()
	}()
	if n.Blockchain == nil {
		n.SetBlockchain(blockchain.ContinueBlockChain(n.GetAddress()))
	}
	return n.Blockchain.AddTSXMemPool(tsx)
}
func (n *FullNode) GetMemPool() *blockchain.Transaction {
	return n.Blockchain.GetTSXMemPool()
}

func (n *FullNode) DeleteTSXMemPool(ID []byte) {
	n.Blockchain.DeleteTSXMemPool(ID)
}
func (s *BaseNode) GetBlockChain(ctx context.Context, re *pb.Empty) (*pb.Response, error) {
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
	}()

	dbMutex.Lock()
	defer dbMutex.Unlock()
	bc := blockchain.ContinueBlockChain(s.Address)
	defer bc.Database.Close()

	Blocks := bc.GetBlocks()
	var blocks []*pb.Block
	for _, block := range Blocks {
		var transactions []*pb.Transaction
		for _, tx := range block.Transactions {
			var Vin []*pb.TXInput
			for _, input := range tx.Vin {
				Vin = append(Vin, &pb.TXInput{
					Txid:      input.Txid,
					Vout:      int32(input.Vout),
					Signature: input.Signature,
					PubKey:    input.PubKey,
				})
			}
			var Vout []*pb.TXOutput
			for _, output := range tx.Vout {
				Vout = append(Vout, &pb.TXOutput{
					Value:      int32(output.Value),
					PubKeyHash: output.PubKeyHash,
				})
			}
			transactions = append(transactions, &pb.Transaction{
				Id:   tx.ID,
				Vin:  Vin,
				Vout: Vout,
			})
		}
		blocks = append(blocks, &pb.Block{
			Timestamp:    block.TimeStamp.Unix(),
			Hash:         block.Hash,
			Transactions: transactions,
			PrevHash:     block.PrevHash,
			Nonce:        int32(block.Nonce),
		})
	}
	response := &pb.Response{
		Blocks:  blocks,
		Status:  stateName[s.Status],
		Success: true,
	}

	return response, nil
}
func (s *FullNode) MineBlock(ctx context.Context, in *pb.Empty) (*pb.Response, error) {
	s.Status = StateMining
	defer func() {
		s.Status = StateIdle
		s.CloseDB()
	}()
	s.SetBlockchain(blockchain.ContinueBlockChain(s.GetAddress()))
	dbMutex.Lock()
	defer dbMutex.Unlock()
	tsx := s.GetMemPool()
	if tsx == nil {
		return &pb.Response{Success: false}, nil
	}
	block, err := s.GetBlockchain().AddBlock([]*blockchain.Transaction{tsx})
	if err != true {
		return &pb.Response{Success: false}, nil
	}
	UTXOSet := blockchain.UTXOSet{BlockChain: s.GetBlockchain()}
	UTXOSet.Update(block)
	s.GetBlockchain().DeleteTSXMemPool(tsx.ID)
	return &pb.Response{Success: true}, nil
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
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
	}()
	var wg sync.WaitGroup
	for _, peer := range s.FullPeers {
		wg.Add(1)
		go func(peer *FullNode) {
			defer wg.Done()
			fmt.Printf("Sending block to %s\n", peer.GetAddress())
			conn, err := grpc.NewClient(peer.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", peer.GetAddress(), err)
				return
			}
			defer conn.Close()
			client := pb.NewBlockchainServiceClient(conn)
			added, err := client.AddBlock(ctx, block)
			if added.Success {
				fmt.Printf("Block sent to %s\n", peer.GetAddress())
				peer.BroadcastBlock(ctx, block)
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
	responseChan := make(chan *pb.Response, 1)
	errChan := make(chan error, 1)

	go func() {
		defer close(responseChan)
		defer close(errChan)

		s.Status = StateConnected
		defer func() {
			s.Status = StateIdle
		}()

		errorCount := int32(0)
		successCount := int32(0)
		var wg sync.WaitGroup

		for _, peer := range s.FullPeers {
			wg.Add(1)
			go func(peer *FullNode) {
				defer wg.Done()

				select {
				case <-ctx.Done():
					atomic.AddInt32(&errorCount, 1)
					return
				default:
					if err := s.broadcastToPeer(ctx, peer, in); err != nil {
						atomic.AddInt32(&errorCount, 1)
						fmt.Printf("Broadcast to %s failed: %v\n", peer.GetAddress(), err)
					} else {
						atomic.AddInt32(&successCount, 1)
					}
				}
			}(peer)
		}

		wg.Wait()

		if errorCount == int32(len(s.FullPeers)) {
			errChan <- fmt.Errorf("broadcast failed to all peers")
			return
		}

		responseChan <- &pb.Response{
			Status:  stateName[s.Status],
			Success: successCount > 0,
		}
	}()

	return <-responseChan, <-errChan
}

// @input: ctx context.Context - context for the operation
// @input: peer *FullNode - peer to broadcast to
// @input: in *pb.Transaction - transaction to broadcast
// @return: error - error if any occurred during broadcast
func (s *BaseNode) broadcastToPeer(ctx context.Context, peer *FullNode, in *pb.Transaction) error {
	conn, err := grpc.NewClient(
		peer.GetAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)

	if _, err = client.AddTSXMempool(ctx, in); err != nil {
		return fmt.Errorf("failed to AddTSXMempool: %v", err)
	}
	// FIXME(12): wrong implementation for the floading algo
	// expected to recusively broadcast to all peers except the one that sent the transaction to the current node
	for _, p := range peer.FullPeers {
		if err := s.broadcastToPeer(ctx, p, in); err != nil {
			return err
		}
	}
	return nil
}
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

	if in.Type == "full" {
		s.FullPeers[in.Address] = &FullNode{
			BaseNode: BaseNode{
				ID:      in.NodeId,
				Address: in.Address,
				Status:  getStatusFromString(in.Status),
				Type:    in.Type,
				Wallets: mapServerWalletsToBlockchainWallets(in.Wallets),
				Blockchain: &blockchain.BlockChain{
					LastTimeUpdate: in.LastTimeUpdate,
					Height:         int(in.Height),
				},
			},
		}
	} else if in.Type == "wallet" {
		s.WalletPeers[in.Address] = &WalletNode{
			BaseNode: BaseNode{
				ID:      in.NodeId,
				Address: in.Address,
				Status:  getStatusFromString(in.Status),
				Type:    in.Type,
				Wallets: mapServerWalletsToBlockchainWallets(in.Wallets),
				Blockchain: &blockchain.BlockChain{
					LastTimeUpdate: in.LastTimeUpdate,
					Height:         int(in.Height),
				},
			},
		}
	}
	return &pb.AddPeerResponse{Success: true}, nil
}
func (s *BaseNode) AddBlock(ctx context.Context, in *pb.Block) (*pb.Response, error) {
	s.Status = StateMining
	defer func() {
		s.Status = StateIdle
		s.Blockchain.Database.Close()
	}()
	fmt.Println("Received new block")
	dbMutex.Lock()
	defer dbMutex.Unlock()
	block := &blockchain.Block{
		TimeStamp: time.Unix(in.Timestamp, 0),
		Hash:      in.Hash,
		PrevHash:  in.PrevHash,
		Nonce:     int(in.Nonce),
	}
	transactions := make([]*blockchain.Transaction, 0, len(in.Transactions))
	for _, tx := range in.Transactions {
		transaction := mapServerTsxToBlockchainTsx(tx)
		transactions = append(transactions, transaction)
	}
	block.Transactions = transactions
	proof := blockchain.NewProof(block)
	if !proof.Validate() {
		return &pb.Response{Success: false}, nil
	}
	if err := s.Blockchain.AddEnireBlock(block); !err {
		return &pb.Response{Success: false}, nil
	}
	UTXOSet := blockchain.UTXOSet{BlockChain: s.Blockchain}
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
			ID:          int32(id),
			Address:     address,
			Blockchain:  blockchain.ContinueBlockChain(address),
			Status:      StateIdle,
			FullPeers:   make(map[string]*FullNode),
			WalletPeers: make(map[string]*WalletNode),
			Type:        "full",
			Wallets:     wallets,
		},
	}
}

func NewWalletNode() *WalletNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	return &WalletNode{
		BaseNode: BaseNode{
			ID:          int32(id),
			Address:     address,
			Blockchain:  blockchain.ContinueBlockChain(address),
			Status:      StateIdle,
			FullPeers:   make(map[string]*FullNode),
			WalletPeers: make(map[string]*WalletNode),
			Type:        "wallet",
			Wallets:     wallets,
		},
	}
}

func (n *BaseNode) DiscoverPeers() (bestPeer *BaseNode) {

	// Step 1: Gather metadata from each peer
	for _, port := range availablePorts {
		fullAddress := "localhost:" + port
		if fullAddress != n.Address {
			conn, err := grpc.NewClient(fullAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Printf("Failed to connect to %s\n", fullAddress)
				continue
			}
			defer conn.Close()

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
				bestPeer = &currentPeer.BaseNode
			} else if currentPeer.BaseNode.Blockchain != nil &&
				bestPeer.Blockchain != nil &&
				bestPeer.Blockchain.GetHeight() < currentPeer.BaseNode.Blockchain.GetHeight() {
				bestPeer = &currentPeer.BaseNode
			}
		}
	}
	return bestPeer
}

// @input: peer *BaseNode - peer node to fetch blockchain from
// @return: []*pb.Block - array of blocks fetched from peer
// @return: error - error if any occurred during fetch
func (n *BaseNode) FetchBlockchain(peer *BaseNode) ([]*pb.Block, error) {
	if peer == nil || peer.GetAddress() == n.GetAddress() {
		return nil, nil
	}
	fmt.Printf("Fetching blockchain from peer %s\n", peer.GetAddress())
	conn, err := grpc.NewClient(
		peer.GetAddress(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to peer %s: %v", peer.GetAddress(), err)
	}
	defer conn.Close()

	client := pb.NewBlockchainServiceClient(conn)
	response, err := client.GetBlockChain(context.Background(), &pb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch blockchain from peer %s: %v", peer.GetAddress(), err)
	}

	fmt.Printf("Fetched blockchain from peer %s with %d blocks\n",
		peer.GetAddress(), len(response.Blocks))
	return response.Blocks, nil
}

// @input: newBlocks []*pb.Block - array of new blocks to be added to blockchain
// @return: error - error if any occurred during update
func (n *BaseNode) UpdateBlockchain(newBlocks []*pb.Block) error {
	if n.Blockchain == nil || n.Blockchain.Database == nil {
		n.Blockchain = blockchain.ContinueBlockChain(n.Address)
	}
	newBlocks = newBlocks[n.Blockchain.GetHeight():]
	prevHeight := n.Blockchain.GetHeight()
	for _, protoBlock := range newBlocks {
		transactions := make([]*blockchain.Transaction, 0, len(protoBlock.Transactions))
		for _, tx := range protoBlock.Transactions {
			transaction := convertProtoTransaction(tx)
			transactions = append(transactions, transaction)
		}
		// check the prev hash of the new block with the last block in the blockchain
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
		if err := n.Blockchain.AddEnireBlock(block); !err {
			fmt.Println("Failed to add block to blockchain")
			continue
		}
	}
	if n.Blockchain != nil && prevHeight < n.Blockchain.GetHeight() {
		fmt.Printf("Blockchain updated to height %d\n", n.Blockchain.GetHeight())
		UTXOSet := blockchain.UTXOSet{BlockChain: n.Blockchain}
		UTXOSet.Reindex()
	}
	return nil
}

// @input: tx *pb.Transaction - protobuf transaction to be converted
// @return: *blockchain.Transaction - converted internal transaction
func convertProtoTransaction(tx *pb.Transaction) *blockchain.Transaction {
	inputs := make([]blockchain.TXInput, 0, len(tx.Vin))
	for _, input := range tx.Vin {
		inputs = append(inputs, blockchain.TXInput{
			Txid:      input.Txid,
			Vout:      int(input.Vout),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}

	outputs := make([]blockchain.TXOutput, 0, len(tx.Vout))
	for _, output := range tx.Vout {
		outputs = append(outputs, blockchain.TXOutput{
			Value:      int(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}

	return &blockchain.Transaction{
		ID:   tx.Id,
		Vin:  inputs,
		Vout: outputs,
	}
}

// @input: n Node - node to sync blockchain for
// @input: bestPeer *BaseNode - peer to sync blockchain from
// @return: error - error if any occurred during sync
func SyncBlockchain(n Node, bestPeer *BaseNode) error {
	defer n.GetBlockchain().Database.Close()
	blocks, err := n.FetchBlockchain(bestPeer)
	if blocks == nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to fetch blockchain: %v", err)
	}

	if err := n.UpdateBlockchain(blocks); err != nil {
		return fmt.Errorf("failed to update blockchain: %v", err)
	}

	if n.GetBlockchain() == nil || n.GetBlockchain().Database == nil {
		fmt.Println("No existing blockchain found! Initializing new blockchain.")
		n.SetBlockchain(blockchain.ContinueBlockChain(n.GetAddress()))
	}

	n.GetBlockchain().Print()
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
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080") // Allow Vue.js frontend
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
			n.CloseDB()
		}()
		pairs := n.GetAllPeers()
		response, _ := json.Marshal(pairs)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			n.CloseDB()
		}()
		response, _ := json.Marshal(n)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			n.CloseDB()
		}()
		response, _ := json.Marshal(stateName[n.GetStatus()])
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		fromWallet := r.FormValue("from")
		toWallet := r.FormValue("to")
		amountStr := r.FormValue("amount")
		amount := util.StrToInt(amountStr)
		defer func() {
			n.CloseDB()
		}()
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
		bc := blockchain.ContinueBlockChain(n.GetAddress())
		UTXO := blockchain.UTXOSet{
			BlockChain: bc,
		}
		tx := ws.NewTransaction(fromWallet, toWallet, amount, &UTXO, n.GetAddress())
		if tx == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Transaction failed"))
		}
		if fullNode, ok := n.(*FullNode); ok {
			syncMemPool(*fullNode, tx)
		}
		tsx := mapBlockchainTsxToServerTsx(tx)
		n.BroadcastTSX(context.Background(), tsx)
		miningNode := SelectRandomFullNode(n)
		if miningNode != nil {
			fmt.Println("Mining block with node: ", miningNode.GetAddress())
			conn, err := grpc.NewClient(miningNode.GetAddress(), grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
			if err != nil {
				fmt.Println("Error creating client: ", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error creating client"))
			}
			defer conn.Close()
			client := pb.NewBlockchainServiceClient(conn)
			_, err = client.MineBlock(context.Background(), &pb.Empty{})
			if err != nil {
				fmt.Printf("Error mining block: %v\n", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Error mining block"))
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Transaction sent"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("No mining node available"))
		}
	})
	mux.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			n.CloseDB()
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
	s := &http.Server{
		Handler: enableCORS(mux),
	}
	return s.Serve(l)
}
func syncMemPool(n FullNode, tx *blockchain.Transaction) {
	n.AddMemPool(tx)
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
