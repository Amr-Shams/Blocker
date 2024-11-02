package node

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"log"
	"net"
	"net/http"
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
	"golang.org/x/net/context"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// ADD(7): a new presentation for the project using HTMX and TailwindCSS
// ADD(8): new feature
// for the clock sync between the nodes(version vector clock, logical clock)
const (
	StateIdle = iota
	StateMining
	StateValidating
	StateConnected
)

var stateName = map[int]string{
	StateIdle:      "Idle",
	StateMining:    "Mining",
	StateConnected: "Connected",
}

type Node interface {
	GetBlockChain(context.Context, *pb.Empty) (*pb.Response, error)
	BroadcastBlock(context.Context, *pb.Transaction) (*pb.Response, error)
	AddPeer(context.Context, *pb.AddPeerRequest) (*pb.AddPeerResponse, error)
	GetPeers(context.Context, *pb.Empty) (*pb.GetPeersResponse, error)
	GetAddress() string
	CheckStatus(context.Context, *pb.Request) (*pb.Response, error)
	SetStatus(int)
	GetBlockchain() *blockchain.BlockChain
	SetBlockchain(*blockchain.BlockChain)
	GetPort() string
	AddBlock(context.Context, *pb.Transaction) (*pb.Response, error)
	Hello(context.Context, *pb.Request) (*pb.HelloResponse, error)
	AllPeersAddress() []string
	DiscoverPeers()
	FetchAndMergeBlockchains() []*pb.Block
	UpdateBlockchain([]*pb.Block)
	GetWallets() *wallet.Wallets
    AddWallet(*wallet.Wallet)
}
type BaseNode struct {
	ID         int32
	Address    string
	Status     int
	Blockchain *blockchain.BlockChain
	Peers      map[string]*BaseNode
	Type       string
	LastTime   time.Time
	Wallets    *wallet.Wallets
}

type FullNode struct {
	BaseNode
}

type WalletNode struct {
	BaseNode
}

var dbMutex sync.Mutex

func (s *BaseNode) GetBlockChain(ctx context.Context, re *pb.Empty) (*pb.Response, error) {
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
	}()

	dbMutex.Lock()
	defer dbMutex.Unlock()

	if s.GetBlockchain() != nil && s.GetBlockchain().Database != nil {
		s.GetBlockchain().Database.Close()
	}

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

func MergeBlockChain(blockchain1, blockchain2 []*pb.Block) []*pb.Block {
	var mergedBlockChain []*pb.Block
	addedBlocks := make(map[string]bool)
	for _, block := range blockchain2 {
		mergedBlockChain = append(mergedBlockChain, block)
		addedBlocks[string(block.Hash)] = true
	}
	for _, block := range blockchain1 {
		if _, ok := addedBlocks[string(block.Hash)]; !ok {
			mergedBlockChain = append(mergedBlockChain, block)
		} else {
			mergedBlockChain = append(mergedBlockChain, resolveConflicts(block, block))
		}
	}
	return mergedBlockChain
}
func resolveConflicts(block1, block2 *pb.Block) *pb.Block {
	if block1.Timestamp > block2.Timestamp {
		return block1
	} else {
		return block2
	}
}

func (s *BaseNode) BroadcastBlock(ctx context.Context, in *pb.Transaction) (*pb.Response, error) {
	s.Status = StateConnected
	defer func() {
		s.Status = StateIdle
	}()
	var wg sync.WaitGroup
	for _, peer := range s.Peers {
		wg.Add(1)
		go func(peer *BaseNode) {
			defer wg.Done()
			fmt.Printf("Sending block to %s\n", peer.GetAddress())
			conn, err := grpc.NewClient(peer.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				fmt.Printf("Failed to connect to %s: %v\n", peer.GetAddress(), err)
				return
			}
			defer conn.Close()
			client := pb.NewBlockchainServiceClient(conn)
			added, err := client.AddBlock(ctx, in)
			if err != nil {
				fmt.Printf("Failed to send block to %s: %v\n", peer.GetAddress(), err)
				return
			}
			if added.Success {
				fmt.Printf("Block sent to %s\n", peer.GetAddress())
				peer.BroadcastBlock(ctx, in)
			} else {
				fmt.Printf("Block not sent to %s\n", peer.GetAddress())
			}
		}(peer)
	}
	wg.Wait()
	return &pb.Response{Status: stateName[s.Status], Success: true}, nil
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
	s.LastTime = time.Now()
}
func (s *BaseNode) GetPort() string {
	return s.Address[strings.LastIndex(s.Address, ":")+1:]
}
func (s *BaseNode) AddWallet(w *wallet.Wallet) {
    s.Wallets.Wallets[string(w.PublicKey)] = w
}
func (s *BaseNode) AddPeer(ctx context.Context, in *pb.AddPeerRequest) (*pb.AddPeerResponse, error) {
	s.Peers[in.Address] = &BaseNode{
		ID:      in.NodeId,
		Address: in.Address,
		Status:  getStatusFromString(in.Status),
		Type:    in.Type,
	}
	return &pb.AddPeerResponse{Success: true}, nil
}
func (s *BaseNode) AddBlock(ctx context.Context, in *pb.Transaction) (*pb.Response, error) {
	s.Status = StateMining
	defer func() {
		s.Status = StateIdle
	}()
	fmt.Println("Received new block")

	Vin := []blockchain.TXInput{}
	for _, input := range in.Vin {
		Vin = append(Vin, blockchain.TXInput{
			Txid:      input.Txid,
			Vout:      int(input.Vout),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}
	Vout := []blockchain.TXOutput{}
	for _, output := range in.Vout {
		Vout = append(Vout, blockchain.TXOutput{
			Value:      int(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}
	tsx := &blockchain.Transaction{
		ID:   in.Id,
		Vin:  Vin,
		Vout: Vout,
	}

	if s.GetBlockchain() == nil || s.GetBlockchain().Database == nil {
		s.SetBlockchain(blockchain.ContinueBlockChain(s.GetAddress()))
	}
	block, added := s.GetBlockchain().AddBlock([]*blockchain.Transaction{tsx})
	if !added {
		s.Status = StateConnected
		return &pb.Response{Status: stateName[s.Status], Success: false}, nil
	}

	fmt.Printf("Added new block with hash: %x\n", block.Hash)
	return &pb.Response{Status: stateName[s.Status], Success: true}, nil
}
func mapServerWalletToBlockchainWallet(w *pb.Wallet) *wallet.Wallet {
	block, _ := pem.Decode([]byte(w.PrivateKey))
	if block == nil || block.Type != "EC PRIVATE KEY" {
		log.Fatalf("Failed to decode PEM block containing private key")
	}
	priv, err := x509.ParseECPrivateKey(block.Bytes)
	util.Handle(err)
	return &wallet.Wallet{
		PrivateKey: *priv,
		PublicKey:  w.PublicKey,
	}
}
func mapBlockchainWalletToServerWallet(wallet *wallet.Wallet) *pb.Wallet {
	if wallet == nil {
        return &pb.Wallet{}
    }
    privBytes, err := x509.MarshalECPrivateKey(&wallet.PrivateKey)
    util.Handle(err)
	peemPriv := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: privBytes})
	return &pb.Wallet{
		PrivateKey: peemPriv,
		PublicKey:  wallet.PublicKey,
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
    if wallets == nil  || wallets.Wallets == nil { 
        return ws
    }
	for _, w := range wallets.Wallets {
		ws.Wallets = append(ws.Wallets, mapBlockchainWalletToServerWallet(w))
	}
	return ws
}
func (s *BaseNode) Hello(ctx context.Context, in *pb.Request) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		NodeId: s.ID, Address: s.Address, Status: stateName[s.Status], Type: s.Type, LastHash: s.Blockchain.LastHash, LastUpdate: s.LastTime.Unix(),
		Wallets: mapBlockchainWalletsToServerWallets(s.GetWallets()),
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
	for _, peer := range s.Peers {
		response = append(response, &pb.HelloResponse{
			NodeId:     peer.ID,
			Address:    peer.Address,
			Status:     stateName[peer.Status],
			Type:       peer.Type,
			LastHash:   peer.GetLastHash(),
			LastUpdate: peer.LastTime.Unix(),
			Wallets:    mapBlockchainWalletsToServerWallets(peer.GetWallets()),
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
	for _, peer := range b.Peers {
		peers = append(peers, peer.GetAddress())
	}
	return peers
}

func NewFullNode() *FullNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	return &FullNode{
		BaseNode: BaseNode{
			ID:         int32(id),
			Address:    address,
			Blockchain: blockchain.ContinueBlockChain(address),
			Status:     StateIdle,
			Peers:      make(map[string]*BaseNode),
			Type:       "full",
			Wallets:    wallets,
		},
	}
}

func NewWalletNode() *WalletNode {
	id := uuid.New().ClockSequence()
	address := viper.GetString("NodeID")
	wallets := wallet.LoadWallets(address)
	return &WalletNode{
		BaseNode: BaseNode{
			ID:         int32(id),
			Address:    address,
			Blockchain: blockchain.ContinueBlockChain(address),
			Status:     StateIdle,
			Peers:      make(map[string]*BaseNode),
			Type:       "wallet",
			Wallets:    wallets,
		},
	}
}

func (n *BaseNode) DiscoverPeers() {
	n.Peers = make(map[string]*BaseNode)
	portStart := 5001
	portEnd := 5004

	for port := portStart; port <= portEnd; port++ {
		fullAddress := fmt.Sprintf("localhost:%d", port)
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
			fmt.Printf("Peers Lenght in client %d\n", len(peers.Peers))
            fmt.Println("Current node wallets length ", len(n.GetWallets().Wallets))
			_, err = client.AddPeer(context.Background(), &pb.AddPeerRequest{
				NodeId:  int32(n.ID),
				Address: n.GetAddress(),
				Status:  stateName[n.Status],
				Type:    n.Type,
				Wallets: mapBlockchainWalletsToServerWallets(n.GetWallets()),
			})
			if err != nil {
				fmt.Printf("Failed to add peer %s: %v\n", fullAddress, err)
				continue
			}
			n.Peers[fullAddress] = &BaseNode{
				ID:       peerResponse.NodeId,
				Address:  peerResponse.Address,
				Status:   getStatusFromString(peerResponse.Status),
				Type:     peerResponse.Type,
				LastTime: time.Unix(peerResponse.LastUpdate, 0),
				Wallets:  mapServerWalletsToBlockchainWallets(peerResponse.Wallets),
			}
			peers, _ = client.GetPeers(context.Background(), &pb.Empty{})
			fmt.Printf("Peers Lenght in client %d\n", len(peers.Peers))
		}
	}
}
func (n *BaseNode) FetchAndMergeBlockchains() []*pb.Block {
	var wg sync.WaitGroup
	fetchedBlocks := make(chan *pb.Response)
	mergedBlockChain := []*pb.Block{}

	go func() {
		var mu sync.Mutex
		for blockchain := range fetchedBlocks {
			if blockchain != nil {
				mu.Lock()
				mergedBlockChain = MergeBlockChain(mergedBlockChain, blockchain.Blocks)
				mu.Unlock()
			}
		}
	}()

	for _, peer := range n.Peers {
		if peer.GetAddress() == n.GetAddress() {
			continue
		}
		wg.Add(1)
		go func(peer *BaseNode) {
			defer wg.Done()
			conn, err := grpc.NewClient(peer.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			util.Handle(err)
			defer conn.Close()
			client := pb.NewBlockchainServiceClient(conn)
			blockchain, err := client.GetBlockChain(context.Background(), &pb.Empty{})
			util.Handle(err)
			fetchedBlocks <- blockchain
		}(peer)
	}
	wg.Wait()
	close(fetchedBlocks)
	return mergedBlockChain
}

func (n *BaseNode) UpdateBlockchain(mergedBlockChain []*pb.Block) {
	if n.Blockchain == nil || n.Blockchain.Database == nil {
		n.Blockchain = blockchain.ContinueBlockChain(n.Address)
	}
	var updateUtxo = false
	for i := len(mergedBlockChain) - 1; i >= 0; i-- {
		b := mergedBlockChain[i]
		tsx := []*blockchain.Transaction{}
		for _, tx := range b.Transactions {
			Vin := []blockchain.TXInput{}
			for _, input := range tx.Vin {
				Vin = append(Vin, blockchain.TXInput{
					Txid:      input.Txid,
					Vout:      int(input.Vout),
					Signature: input.Signature,
					PubKey:    input.PubKey,
				})
			}
			Vout := []blockchain.TXOutput{}
			for _, output := range tx.Vout {
				Vout = append(Vout, blockchain.TXOutput{
					Value:      int(output.Value),
					PubKeyHash: output.PubKeyHash,
				})
			}
			tsx = append(tsx, &blockchain.Transaction{
				ID:   tx.Id,
				Vin:  Vin,
				Vout: Vout,
			})
		}
		block := &blockchain.Block{
			TimeStamp:    time.Unix(b.Timestamp, 0),
			Hash:         b.Hash,
			Transactions: tsx,
			PrevHash:     b.PrevHash,
			Nonce:        int(b.Nonce),
		}
		proof := blockchain.NewProof(block)
		if !proof.Validate() {
			continue
		}
		if added := n.Blockchain.AddEnireBlock(block); added {
			updateUtxo = true
		}
	}
	if updateUtxo {
		UTXO := blockchain.UTXOSet{n.Blockchain}
		UTXO.Reindex()
	}
}

func grpcServer(l net.Listener, n Node) error {
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	pb.RegisterBlockchainServiceServer(grpcServer, n)
	return grpcServer.Serve(l)
}
func (n *BaseNode) GetAllPairs() map[string]*BaseNode {
	return n.Peers
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
func httpServer(l net.Listener, n *WalletNode) error {
	mux := http.NewServeMux()
	// create a get request to get all paris
	// create a get request to get info about the node

	mux.HandleFunc("/pairs", func(w http.ResponseWriter, r *http.Request) {
		pairs := n.GetAllPairs()
		response, _ := json.Marshal(pairs)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		response, _ := json.Marshal(n)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		response, _ := json.Marshal(stateName[n.Status])
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
		UTXO := blockchain.UTXOSet{n.GetBlockchain()}
		tx := ws.NewTransaction(fromWallet, toWallet, amount, &UTXO, n.Address)
		tsx := mapBlockchainTsxToServerTsx(tx)
		block, added := n.GetBlockchain().AddBlock([]*blockchain.Transaction{tx})
		if !added {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to add block"))
			return
		}
		UTXO.Update(block)
		n.BroadcastBlock(context.Background(), tsx)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Transaction sent"))
	})
	mux.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
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

func StartWalletNodeCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "run2",
		Short:   "Start a wallet node",
		Long:    "a wallet node that can send and receive transactions",
		Example: "wallet print",
		Args:    cobra.ExactArgs(0),
		Run: func(cmd *cobra.Command, args []string) {
			n := NewWalletNode()
			n.DiscoverPeers()
			fmt.Printf("Node is connected to %d peers\n", len(n.AllPeersAddress()))

			conn, err := grpc.Dial(n.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				log.Fatalf("Failed to create gRPC connection: %v", err)
			}
			defer conn.Close()

			n.SetStatus(StateConnected)
			defer func() {
				n.SetStatus(StateIdle)
			}()

			// Start blockchain synchronization in a goroutine
			go func() {
				if err := syncBlockchain(n); err != nil {
					log.Printf("Error syncing blockchain: %v", err)
				}
			}()

			// Set up server
			port := n.GetPort()
			lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
			if err != nil {
				log.Fatalf("Failed to listen on port %s: %v", port, err)
			}

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

// Helper function to handle blockchain synchronization
func syncBlockchain(n Node) error {
	fmt.Printf("Before fetching blockchain length: %d\n", len(n.GetBlockchain().GetBlocks()))

	mergedBlockChain := n.FetchAndMergeBlockchains()
	n.UpdateBlockchain(mergedBlockChain)
	fmt.Printf("Merged blockchain length: %d\n", len(mergedBlockChain))

	if n.GetBlockchain() == nil || n.GetBlockchain().Database == nil {
		fmt.Println("No existing blockchain found! Initializing new blockchain.")
		n.SetBlockchain(blockchain.ContinueBlockChain(n.GetAddress()))
	}

	n.GetBlockchain().Print()
	fmt.Printf("Chain of Node %s has %d blocks\n", n.GetAddress(), len(n.GetBlockchain().GetBlocks()))

	defer n.GetBlockchain().Database.Close()
	return nil
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
			n.DiscoverPeers()
			fmt.Printf("Node is connected to %d peers\n", len(n.AllPeersAddress()))

			conn, err := grpc.NewClient(n.GetAddress(), grpc.WithTransportCredentials(insecure.NewCredentials()))
			util.Handle(err)
			defer conn.Close()

			n.SetStatus(StateConnected)
			defer func() {
				n.SetStatus(StateIdle)
			}()
			go func() {
				if err := syncBlockchain(n); err != nil {
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
			//httpL := m.Match(cmux.HTTP1Fast())

			g := new(errgroup.Group)

			g.Go(func() error {
				log.Printf("Starting gRPC server on port %s\n", port)
				if err := grpcServer(grpcL, n); err != nil {
					return fmt.Errorf("gRPC server error: %v", err)
				}
				return nil
			})

			/*g.Go(func() error {
				log.Printf("Starting HTTP server on port %s\n", port)
				if err := httpServer(httpL,n); err != nil {
					return fmt.Errorf("HTTP server error: %v", err)
				}
				return nil
			})
			*/

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
