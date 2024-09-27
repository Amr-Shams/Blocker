package blockchain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/Amr-Shams/Blocker/util"
	"github.com/cbergoon/merkletree"
)

type Block struct {
	TimeStamp    time.Time
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte
	Nonce        int
}

func NewBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{TimeStamp: time.Now(), Hash: []byte{}, Transactions: txs, PrevHash: prevHash, Nonce: 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func Genesis(Coinbase *Transaction) *Block {
	return NewBlock([]*Transaction{Coinbase}, []byte{})
}
func (b *Block) HashTransactions() []byte {
	var txHashes []merkletree.Content
	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx)
	}
	mTree, err := merkletree.NewTree(txHashes)
	util.Handle(err)
	return mTree.MerkleRoot()
}
func (b *Block) Serialize() []byte {
	return util.Serialize(b)
}

func (b *Block) Deserialize(data []byte) {
	util.Deserialize(data, b)
}

func (b *Block) Print() {
	fmt.Printf("TimeStamp: %s\n", b.TimeStamp)
	fmt.Printf("Hash: %x\n", b.Hash)
	for _, tx := range b.Transactions {
		fmt.Println(tx)
	}
	fmt.Println("PrevHash:", b.PrevHash)
	fmt.Println("Nonce:", b.Nonce)
	pow := NewProof(b)
	fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
}
