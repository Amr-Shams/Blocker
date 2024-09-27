package node

import (
	"context"
	"testing"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/wallet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
)

// Mocking the BlockchainServiceClient
type MockBlockchainServiceClient struct {
	mock.Mock
}

func (m *MockBlockchainServiceClient) GetBlockChain(ctx context.Context, in *pb.Request, opts ...grpc.CallOption) (*pb.Response, error) {
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

func TestNode_GetBlockChain(t *testing.T) {
	node := &Node{
		ID:      1,
		Address: "5001",
		Peers:   make(map[string]*Node),
		Status:  StateIdle,
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

	node.Peers["localhost:5002"] = &Node{Address: "localhost:5002"}
	node.Peers["localhost:5003"] = &Node{Address: "localhost:5003"}

	ctx := context.Background()
	response, err := node.GetBlockChain(ctx, &pb.Request{})

	assert.NoError(t, err)
	assert.Equal(t, 2, len(response.Blocks))
	assert.Equal(t, stateName[StateIdle], response.Status)
	assert.True(t, response.Success)
}

func TestNode_BroadcastBlock(t *testing.T) {
	node := &Node{
		ID:      1,
		Address: "localhost:5001",
		Peers:   make(map[string]*Node),
		Status:  StateIdle,
	}

	mockClient := new(MockBlockchainServiceClient)
	mockResponse := &pb.Response{
		Status:  stateName[StateConnected],
		Success: true,
	}
	mockClient.On("AddBlock", mock.Anything, mock.Anything).Return(mockResponse, nil)

	node.Peers["localhost:5002"] = &Node{Address: "localhost:5002"}
	node.Peers["localhost:5003"] = &Node{Address: "localhost:5003"}

	ctx := context.Background()
	response, err := node.BroadcastBlock(ctx, &pb.Transaction{})

	assert.NoError(t, err)
	assert.Equal(t, stateName[StateIdle], response.Status)
	assert.True(t, response.Success)
}

func TestNode_AddBlock(t *testing.T) {
	ws := wallet.CreateWallets()
	ws.AddWallet()
	address := ws.GetAllAddresses()[0]
	node := &Node{
		ID:         1,
		Address:    "localhost:5001",
		Peers:      make(map[string]*Node),
		Status:     StateIdle,
		Blockchain: blockchain.NewBlockChain(string(address)),
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

func TestNode_Hello(t *testing.T) {
	node := &Node{
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
