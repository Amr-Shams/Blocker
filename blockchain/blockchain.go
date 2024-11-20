package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"

	"github.com/Amr-Shams/Blocker/storage"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/syndtr/goleveldb/leveldb"
)

const (
	genesis = "First block"
)

var (
	nodeID = "3000"
)

type BlockChain struct {
	LastHash       []byte
	Database       storage.StorageInterface
	LastTimeUpdate int64
	Height         int
}

func (bc *BlockChain) GetLastHash() []byte {
	return bc.LastHash
}
func (bc *BlockChain) GetHeight() int {
	return bc.Height
}
func (bc *BlockChain) GetLastTimeUpdate() int64 {
	return bc.LastTimeUpdate
}

func (bc *BlockChain) AddBlock(txs []*Transaction) (*Block, bool) {
	prevBlock := bc.GetBlock(bc.LastHash)
	if prevBlock == nil {
		util.Handle(fmt.Errorf("previous block not found"))
		return nil, false
	}

	block := NewBlock(txs, prevBlock.Hash)
	bc.LastHash = block.Hash
	bc.LastTimeUpdate = block.TimeStamp.Unix()
	bc.Height++
	err := bc.Database.SaveBlock(block.Hash, block.Serialize())
	if err != nil && bytes.Contains([]byte(err.Error()), []byte("leveldb: key exists")) {
		log.Printf("Block already exists: %x", block.Hash)
		return block, false
	}
	err = bc.Database.SaveLastHash(block.Hash)
	if err != nil {
		util.Handle(err)
		return nil, false
	}
	err = bc.Database.SaveLastTimeUpdate(bc.LastTimeUpdate)
	if err != nil {
		util.Handle(err)
		return nil, false
	}
	err = bc.Database.SaveHeight(bc.Height)
	if err != nil {
		util.Handle(err)
		return nil, false
	}
	return block, true
}

func (bc *BlockChain) GetBlock(blockHash []byte) *Block {
	blockData, err := bc.Database.GetBlock(blockHash)
	if err != nil {
		return nil
	}
	var block Block
	block.Deserialize(blockData)
	return &block
}

func (bc *BlockChain) Print() {
	if bc.Database == nil {
		fmt.Println("No database")
		return
	}
	currentHash := bc.LastHash
	fmt.Println("Last hash: ", hex.EncodeToString(currentHash))
	fmt.Println("Height: ", bc.Height)
	fmt.Println("Last time update: ", bc.LastTimeUpdate)
	for {
		block := bc.GetBlock(currentHash)
		if block == nil {
			fmt.Println("Block not found")
			break
		}
		block.Print()
		if len(block.PrevHash) == 0 {
			break
		}
		currentHash = block.PrevHash
	}
}

func (bc *BlockChain) GetBlocks() []*Block {
	result := []*Block{}
	if bc == nil || bc.LastHash == nil {
		return result
	}
	currentHash := bc.LastHash
	for {
		block := bc.GetBlock(currentHash)
		if block == nil {
			break
		}
		result = append([]*Block{block}, result...)
		if len(block.PrevHash) == 0 {
			break
		}
		currentHash = block.PrevHash
	}
	return result
}

func NewBlockChain(address string, nodeId string) *BlockChain {
	var lastHash []byte
	db, err := storage.GetInstance(nodeId)
	nodeID = nodeId
	lastTimeUpdate := int64(0)
	height := 0
	util.Handle(err)
	lastHash, err = db.GetLastHash()
	if err != nil && address != "" {
		if err == leveldb.ErrNotFound {
			cbtx := NewCoinbaseTX(address, genesis)
			genesis := Genesis(cbtx)
			lastHash = genesis.Hash
			lastTimeUpdate = genesis.TimeStamp.Unix()
			height = 1
			err = db.SaveBlock(genesis.Hash, genesis.Serialize())
			util.Handle(err)
			err = db.SaveLastHash(genesis.Hash)
			util.Handle(err)
			err = db.SaveLastTimeUpdate(lastTimeUpdate)
			util.Handle(err)
			err = db.SaveHeight(height)
			util.Handle(err)
		} else {
			return &BlockChain{nil, db, 0, 0}
		}
	}
	return &BlockChain{lastHash, db, lastTimeUpdate, height}
}

func ContinueBlockChain(nodeId string) *BlockChain {
	db, err := storage.GetInstance(nodeId)
	nodeID = nodeId
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	lastHash, err := db.GetLastHash()
	if err == leveldb.ErrNotFound {
		fmt.Println("No last hash found, initializing new blockchain")
		return NewBlockChain("", nodeId)
	} else if err != nil {
		log.Fatalf("Failed to get last hash: %v", err)
	}
	lastTimeUpdate, err := db.GetLastTimeUpdate()
	util.Handle(err)
	height, err := db.GetHeight()
	util.Handle(err)
	return &BlockChain{lastHash, db, lastTimeUpdate, int(height)}
}
func (bc *BlockChain) Close() {
	bc.Database.Close()
}
func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	iter := bc.Database.NewIterator()
	for {
		iter.Next()
		var block Block
		blockData := iter.Value()
		block.Deserialize(blockData)
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return unspentTXs
}
func (bc *BlockChain) FindUTXO() map[string]TXOutputs {
	UTXOs := make(map[string]TXOutputs)
	spentUTXOs := make(map[string][]int)
	iter := bc.Database.NewIterator()
	for {
		iter.Next()
		var block Block
		block.Deserialize(iter.Value())
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Vout {
				if spentUTXOs[txID] != nil {
					for _, spentOut := range spentUTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXOs[txID] = outs
			}
			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inTxID := hex.EncodeToString(in.Txid)
					spentUTXOs[inTxID] = append(spentUTXOs[inTxID], in.Vout)
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXOs
}

func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Database.NewIterator()
	for iter.Next() {
		var block Block
		blockData := iter.Value()
		block.Deserialize(blockData)
		for _, tx := range block.Transactions {
			if bytes.Equal(tx.ID, ID) {
				return *tx, nil
			}
		}
	}
	return Transaction{}, fmt.Errorf("Transaction not found")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)
	for _, in := range tx.Vin {
		prevTX, err := bc.FindTransaction(in.Txid)
		util.Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)
	for _, in := range tx.Vin {
		prevTX, err := bc.FindTransaction(in.Txid)
		util.Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	return tx.Verify(prevTXs)
}

func (bc *BlockChain) AddEntireBlock(block *Block) bool {
	bc.LastHash = block.Hash
	bc.LastTimeUpdate = block.TimeStamp.Unix()
	bc.Height++
	err := bc.Database.SaveBlock(block.Hash, block.Serialize())
	if err != nil && bytes.Contains([]byte(err.Error()), []byte("leveldb: key exists")) {
		log.Printf("Block already exists: %x", block.Hash)
		return false
	}
	util.Handle(err)
	err = bc.Database.SaveLastHash(block.Hash)
	if err != nil {
		util.Handle(err)
		return false
	}
	err = bc.Database.SaveLastTimeUpdate(block.TimeStamp.Unix())
	if err != nil {
		util.Handle(err)
		return false
	}
	err = bc.Database.SaveHeight(bc.Height)
	if err != nil {
		util.Handle(err)
		return false
	}
	return true
}
