package node

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func (s *WalletNode) GetBlockChain(ctx context.Context, re *pb.Empty) (*pb.Response, error) {
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
func (s *WalletNode) BroadcastBlock(ctx context.Context, block *pb.Block) (*pb.Response, error) {
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
			added, _ := client.AddBlock(ctx, block)
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
			added, _ := client.AddBlock(ctx, block)
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

func (s *WalletNode) BroadcastTSX(ctx context.Context, in *pb.Transaction) (*pb.Response, error) {
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

func (s *WalletNode) broadcastToFullNode(ctx context.Context, in *pb.Transaction, peer *FullNode) error {
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

func (s *WalletNode) AddPeer(ctx context.Context, in *pb.AddPeerRequest) (*pb.AddPeerResponse, error) {
	WalletNode := WalletNode{
		ID:      in.NodeId,
		Address: in.Address,
		Status:  getStatusFromString(in.Status),
		Type:    in.Type,
		Wallets: mapServerWalletsToBlockchainWallets(in.Wallets),
		Blockchain: &blockchain.BlockChain{
			LastTimeUpdate: in.LastTimeUpdate,
			Height:         in.Height,
		},
	}
	switch in.Type {
	case "full":
		peer := &FullNode{WalletNode: WalletNode}
		s.FullPeers[in.Address] = peer
		if s.Blockchain.Height < in.Height {
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
		s.WalletPeers[in.Address] = &WalletNode
	default:
		return nil, fmt.Errorf("unknown peer type: %s", in.Type)
	}
	return &pb.AddPeerResponse{Success: true}, nil
}
func (s *WalletNode) AddBlock(ctx context.Context, in *pb.Block) (*pb.Response, error) {
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

func (s *WalletNode) GetPeers(ctx context.Context, in *pb.Empty) (*pb.GetPeersResponse, error) {
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

func (s *WalletNode) Hello(ctx context.Context, in *pb.Request) (*pb.HelloResponse, error) {
	return &pb.HelloResponse{
		NodeId: s.ID, Address: s.Address, Status: stateName[s.Status], Type: s.Type, LastHash: s.Blockchain.LastHash, LastTimeUpdate: s.Blockchain.GetLastTimeUpdate(),
		Wallets: mapBlockchainWalletsToServerWallets(s.GetWallets()), Height: int64(s.Blockchain.GetHeight()),
	}, nil
}

func (s *WalletNode) CheckStatus(ctx context.Context, in *pb.Request) (*pb.Response, error) {
	return &pb.Response{Status: stateName[s.Status], Success: true}, nil
}

func grpcServer(l net.Listener, n serverNode) error {
	grpcServer := grpc.NewServer(grpc.Creds(insecure.NewCredentials()))
	pb.RegisterBlockchainServiceServer(grpcServer, n)
	return grpcServer.Serve(l)
}
