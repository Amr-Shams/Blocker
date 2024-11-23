package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenesis(t *testing.T) {
	tests := []struct {
		name     string
		coinbase *Transaction
	}{
		{
			name:     "Genesis block",
			coinbase: &Transaction{ID: []byte("coinbase"), Vin: nil, Vout: nil},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := Genesis(tt.coinbase)
			assert.NotNil(t, block)
			assert.Equal(t, []*Transaction{tt.coinbase}, block.Transactions)
			assert.Empty(t, block.PrevHash)
			assert.NotEmpty(t, block.Hash)
			assert.NotZero(t, block.Nonce)
		})
	}
}

func TestBlock_HashTransactions(t *testing.T) {
	tests := []struct {
		name string
		txs  []*Transaction
	}{
		{
			name: "Empty transactions",
			txs:  []*Transaction{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewBlock(tt.txs, []byte{})
			hash := block.HashTransactions()
			assert.NotEmpty(t, hash)
		})
	}
}

func TestBlock_SerializeDeserialize(t *testing.T) {
	tests := []struct {
		name string
		txs  []*Transaction
	}{
		{
			name: "Empty transactions",
			txs:  []*Transaction{},
		},
		{
			name: "Single transaction",
			txs: []*Transaction{
				{ID: []byte("tx1"), Vin: nil, Vout: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewBlock(tt.txs, []byte{})
			serialized := block.Serialize()
			assert.NotEmpty(t, serialized)

			var deserializedBlock Block
			deserializedBlock.Deserialize(serialized)
			assert.Equal(t, block.TimeStamp.Unix(), deserializedBlock.TimeStamp.Unix())
			assert.Equal(t, block.Hash, deserializedBlock.Hash)
			assert.Equal(t, block.Transactions, deserializedBlock.Transactions)
			assert.Equal(t, block.PrevHash, deserializedBlock.PrevHash)
			assert.Equal(t, block.Nonce, deserializedBlock.Nonce)
		})
	}
}

func TestBlock_Print(t *testing.T) {
	tests := []struct {
		name string
		txs  []*Transaction
	}{
		{
			name: "Empty transactions",
			txs:  []*Transaction{},
		},
		{
			name: "Single transaction",
			txs: []*Transaction{
				{ID: []byte("tx1"), Vin: nil, Vout: nil},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			block := NewBlock(tt.txs, []byte{})
			block.Print()
		})
	}
}
