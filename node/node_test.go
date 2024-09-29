package node

import (
	"context"
	"log"
	"net"
	"testing"

	"time"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/Amr-Shams/Blocker/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// Mocking the BlockchainServiceClient
type MockBlockchainServiceClient struct {
	mock.Mock
}

func (m *MockBlockchainServiceClient) GetBlockChain(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.Response, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.Response), args.Error(1)
}

func (m *MockBlockchainServiceClient) AddBlock(ctx context.Context, in *pb.Transaction, opts ...grpc.CallOption) (*pb.Response, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.Response), args.Error(1)
}

func (m *MockBlockchainServiceClient) Hello(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.HelloResponse, error) {
	args := m.Called(ctx, in)
	return args.Get(0).(*pb.HelloResponse), args.Error(1)
}

func TestBaseNode_GetBlockChain(t *testing.T) {
	node := &BaseNode{
		ID:         1,
		Address:    "localhost:5001",
		Status:     StateIdle,
		Blockchain: blockchain.ContinueBlockChain("localhost:5001"),
	}

	// Add blocks to the blockchain for testing
	node.Blockchain.AddBlock([]*blockchain.Transaction{
		{ID: []byte("txid1")},
	})
	node.Blockchain.AddBlock([]*blockchain.Transaction{
		{ID: []byte("txid2")},
	})

	ctx := context.Background()
	response, err := node.GetBlockChain(ctx, &pb.Empty{})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Blocks)) // Expecting 2 blocks
	assert.Equal(t, stateName[StateIdle], response.Status)
	assert.True(t, response.Success)
}

func TestBaseNode_BroadcastBlock(t *testing.T) {
	// Start a temporary peer node for testing
	go func() {
		peerNode := &BaseNode{
			ID:      2,
			Address: "localhost:5002",
			Status:  StateIdle,
		}
		lis, err := net.Listen("tcp", peerNode.Address)
		util.Handle(err)
		s := grpc.NewServer()
		pb.RegisterBlockchainServiceServer(s, peerNode)
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	node := &BaseNode{
		ID:         1,
		Address:    "localhost:5001",
		Status:     StateIdle,
		Blockchain: &blockchain.BlockChain{},
		Peers:      map[string]*BaseNode{"localhost:5002": {ID: 2, Address: "localhost:5002", Status: StateIdle}},
	}

	ctx := context.Background()
	tx := &pb.Transaction{
		Id: []byte("txid1"),
		Vin: []*pb.TXInput{
			{Txid: []byte("txid1"), Vout: 0, Signature: []byte("sig1"), PubKey: []byte("pubkey1")},
		},
		Vout: []*pb.TXOutput{
			{Value: 10, PubKeyHash: []byte("pubkeyhash1")},
		},
	}
	response, err := node.BroadcastBlock(ctx, tx)

	assert.NoError(t, err)
	assert.Equal(t, stateName[StateIdle], response.Status)
	assert.True(t, response.Success)
}
func TestBaseNode_AddBlock(t *testing.T) {
	nodeId := "localhost:5001"
	ws := wallet.CreateWallets(nodeId)
	ws.AddWallet(nodeId)
	address := ws.Wallets[nodeId].GetAddress()
	node := &BaseNode{
		ID:         1,
		Address:    "localhost:5001",
		Peers:      make(map[string]*BaseNode),
		Status:     StateIdle,
		Blockchain: blockchain.NewBlockChain(address, nodeId),
	}
	// create a wallet then create a transaction

	mockClient := new(MockBlockchainServiceClient)
	mockResponse := &pb.Response{
		Status:  stateName[StateConnected],
		Success: true,
	}
	mockClient.On("AddBlock", mock.Anything, mock.Anything).Return(mockResponse, nil)

	ctx := context.Background()
	idBytes := []byte("txid1")
	tx := &pb.Transaction{
		Id: idBytes,
		Vin: []*pb.TXInput{
			{Txid: []byte("txid1"), Vout: 0, Signature: []byte("sig1"), PubKey: []byte("pubkey1")},
		},
		Vout: []*pb.TXOutput{
			{Value: 10, PubKeyHash: []byte("pubkeyhash1")},
		},
	}
	response, err := node.AddBlock(ctx, tx)

	assert.NoError(t, err)
	assert.Equal(t, stateName[StateIdle], response.Status)
	assert.True(t, response.Success)
}

func TestBaseNode_Hello(t *testing.T) {
	node := &BaseNode{
		ID:      1,
		Address: "localhost:5001",
		Status:  StateIdle,
	}

	ctx := context.Background()
	response, err := node.Hello(ctx, &pb.Request{})

	assert.NoError(t, err)
	assert.Equal(t, node.ID, response.NodeId)
	assert.Equal(t, node.Address, response.Address)
	assert.Equal(t, stateName[StateIdle], response.Status)
	assert.Nil(t, response.Blocks)
	assert.Empty(t, response.Peers)
}

func TestBaseNode_DiscoverPeers(t *testing.T) {
	node := &BaseNode{
		ID:      1,
		Address: "localhost:5001",
		Status:  StateIdle,
		Peers:   make(map[string]*BaseNode),
	}

	node.DiscoverPeers()

	assert.NotEmpty(t, node.Peers)
	assert.Equal(t, 3, len(node.Peers)) // Assuming there are 3 peers in the range 5001-5004
}

func TestBaseNode_FetchAndMergeBlockchains(t *testing.T) {
	node := &BaseNode{
		ID:      1,
		Address: "localhost:5001",
		Status:  StateIdle,
		Peers:   make(map[string]*BaseNode),
	}

	mockClient := new(MockBlockchainServiceClient)
	mockResponse := &pb.Response{
		Blocks: []*pb.Block{
			{Hash: []byte("block1"), Timestamp: 1},
			{Hash: []byte("block2"), Timestamp: 2},
		},
		Status:  stateName[StateConnected],
		Success: true,
	}
	mockClient.On("GetBlockChain", mock.Anything, mock.Anything).Return(mockResponse, nil)

	node.Peers["localhost:5002"] = &BaseNode{Address: "localhost:5002"}
	node.Peers["localhost:5003"] = &BaseNode{Address: "localhost:5003"}

	mergedBlockChain := node.FetchAndMergeBlockchains()

	assert.NotEmpty(t, mergedBlockChain)
	assert.Equal(t, 2, len(mergedBlockChain))
}

func TestBaseNode_UpdateBlockchain(t *testing.T) {
	nodeID := "localhost:5001"
	ws := wallet.CreateWallets(nodeID)
	ws.AddWallet(nodeID)
	address := ws.Wallets[nodeID].GetAddress()
	node := &BaseNode{
		ID:         1,
		Address:    "localhost:5001",
		Peers:      make(map[string]*BaseNode),
		Status:     StateIdle,
		Blockchain: blockchain.NewBlockChain(address, nodeID),
	}

	mergedBlockChain := []*pb.Block{
		{
			Timestamp: time.Now().Unix(),
			Hash:      []byte("block1"),
			Transactions: []*pb.Transaction{
				{
					Id: []byte("txid1"),
					Vin: []*pb.TXInput{
						{Txid: []byte("txid1"), Vout: 0, Signature: []byte("sig1"), PubKey: []byte("pubkey1")},
					},
					Vout: []*pb.TXOutput{
						{Value: 10, PubKeyHash: []byte("pubkeyhash1")},
					},
				},
			},
			PrevHash: []byte("prevhash1"),
			Nonce:    1,
		},
	}

	node.UpdateBlockchain(mergedBlockChain)

	assert.NotNil(t, node.Blockchain)
	assert.Equal(t, 2, len(node.Blockchain.GetBlocks())) // Assuming the initial blockchain has 1 block
}
