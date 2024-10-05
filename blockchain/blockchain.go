package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"

	"github.com/Amr-Shams/Blocker/util"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const (
	dbFile  = "./tmp/blockchain_%s.data"
	genesis = "First block"
)

type BlockChain struct {
	LastHash []byte
	Database *leveldb.DB
}

func DBExist(nodeId string) bool {
	dbFile := fmt.Sprintf(dbFile, nodeId)
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}
	return true
}

func (bc *BlockChain) AddBlock(txs []*Transaction) (*Block, bool) {
	prevBlock := bc.GetBlock(bc.LastHash)
	if prevBlock == nil {
		util.Handle(fmt.Errorf("previous block not found"))
		return nil, false
	}

	block := NewBlock(txs, prevBlock.Hash)
	err := bc.Database.Put(block.Hash, block.Serialize(), nil)
	if err != nil && bytes.Contains([]byte(err.Error()), []byte("leveldb: key exists")) {
		log.Printf("Block already exists: %x", block.Hash)
		return block, false
	}

	err = bc.Database.Put([]byte("lh"), block.Hash, nil)
	if err != nil {
		util.Handle(err)
		return nil, false
	}

	bc.LastHash = block.Hash
	return block, true
}

func (bc *BlockChain) GetBlock(blockHash []byte) *Block {
	blockData, err := bc.Database.Get(blockHash, nil)
	if err != nil {
		return nil
	}
	var block Block
	block.Deserialize(blockData)
	return &block
}
func (bc *BlockChain) HasBlock(b *Block) (bool, Block) {
	blockData, err := bc.Database.Get(b.Hash, nil)
	if err != nil {
		return false, Block{}
	}
	var block Block
	block.Deserialize(blockData)
	return true, block
}
func (bc *BlockChain) Print() {
	if bc.Database == nil {
		fmt.Println("No database")
		return
	}

	currentHash := bc.LastHash
	fmt.Println("Last hash: ", hex.EncodeToString(currentHash))
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
	fmt.Println("Last hash: ", hex.EncodeToString(currentHash))
	for {
		block := bc.GetBlock(currentHash)
		if block == nil {
			break
		}
		result = append(result, block)
		if len(block.PrevHash) == 0 {
			break
		}
		currentHash = block.PrevHash
	}
	return result
}
func NewEmptyBlockChain(nodeId string) *BlockChain {
	dbFile := fmt.Sprintf(dbFile, nodeId)
	db, err := leveldb.OpenFile(dbFile, &opt.Options{})
	util.Handle(err)
	return &BlockChain{nil, db}
}
func NewBlockChain(address string, nodeId string) *BlockChain {
	var lastHash []byte
	if DBExist(nodeId) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	dbFile := fmt.Sprintf(dbFile, nodeId)
	db, err := leveldb.OpenFile(dbFile, &opt.Options{ErrorIfExist: true})
	util.Handle(err)
	lastHash, err = db.Get([]byte("lh"), nil)
	if err != nil {
		if err == leveldb.ErrNotFound {
			cbtx := NewCoinbaseTX(address, genesis)
			genesis := Genesis(cbtx)
			err = db.Put(genesis.Hash, genesis.Serialize(), nil)
			util.Handle(err)
			err = db.Put([]byte("lh"), genesis.Hash, nil)
			util.Handle(err)
			lastHash = genesis.Hash
		} else {
			util.Handle(err)
		}
	}
	return &BlockChain{lastHash, db}
}

func ContinueBlockChain(nodeId string) *BlockChain {
	if !DBExist(nodeId) {
		fmt.Println("No existing blockchain found!")
		return NewEmptyBlockChain(nodeId)
	}
	dbFile := fmt.Sprintf(dbFile, nodeId)
	fmt.Printf("Continue blockchain from %s\n", dbFile)
	db, err := leveldb.OpenFile(dbFile, &opt.Options{})
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	lastHash, err := db.Get([]byte("lh"), nil)
	if err == leveldb.ErrNotFound {
		fmt.Println("No last hash found, initializing new blockchain")
		db.Close()
		return NewEmptyBlockChain(nodeId)
	} else if err != nil {
		log.Fatalf("Failed to get last hash: %v", err)
	}

	return &BlockChain{lastHash, db}
}
func (bc *BlockChain) Close() {
	bc.Database.Close()
}
func (bc *BlockChain) FindUnspentTransactions(pubKeyHash []byte) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	iter := bc.Database.NewIterator(nil, nil)
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
	iter := bc.Database.NewIterator(nil, nil)
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
	iter := bc.Database.NewIterator(nil, nil)
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

func (bc *BlockChain) AddEnireBlock(b *Block) bool {
	err := bc.Database.Put(b.Hash, b.Serialize(), nil)
	if err != nil && strings.Contains(err.Error(), "leveldb: key") {
		fmt.Println("Block already exists")
		return false
	}
	util.Handle(err)
	bc.LastHash = b.Hash
	err = bc.Database.Put([]byte("lh"), b.Hash, nil)
	util.Handle(err)
	return true
}