package blockchain

import (
	"bytes"
	"encoding/hex"
	"fmt"

	"github.com/Amr-Shams/Blocker/util"
	"github.com/syndtr/goleveldb/leveldb"
)

var (
	utxoPrefix = []byte("utxo-")
	max_number = 1000
)

type UTXOSet struct {
	BlockChain *BlockChain
}

func (UTXO *UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	db := UTXO.BlockChain.Database
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	var accumulated int
	for iter.Next() {
		key := iter.Key()
		if !bytes.HasPrefix(key, []byte(utxoPrefix)) {
			continue
		}
		value := iter.Value()
		key = bytes.TrimPrefix(key, []byte(utxoPrefix))
		var outs TXOutputs
		outs.Deserialize(value)
		txID := hex.EncodeToString(key)
		for outIdx, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}
	return accumulated, unspentOutputs
}
func (UTXO *UTXOSet) FindUTXO(pubKeyHash []byte) []TXOutput {
	db := UTXO.BlockChain.Database
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	var UTXOs []TXOutput
	for iter.Next() {
		key := iter.Key()
		fmt.Printf("key: %x\n", key)
		if !bytes.HasPrefix(key, []byte(utxoPrefix)) {
			fmt.Println("continue")
			continue
		}
		var outs TXOutputs
		outs.Deserialize(iter.Value())
		for _, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}
	return UTXOs
}
func (UTXO *UTXOSet) CountTransactions() int {
	db := UTXO.BlockChain.Database
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	var count int
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), []byte(utxoPrefix)) {
			count++
		}
	}
	return count
}
func (u UTXOSet) Reindex() {
	db := u.BlockChain.Database
	u.DeleteByPrefix()
	UTXO := u.BlockChain.FindUTXO()
	batch := new(leveldb.Batch)
	for txID, outs := range UTXO {
		key, err := hex.DecodeString(txID)
		util.Handle(err)
		key = append(utxoPrefix, key...)
		batch.Put(key, outs.Serialize())
	}
	err := db.Write(batch, nil)
	util.Handle(err)
}

func (u *UTXOSet) Update(block *Block) {
	db := u.BlockChain.Database
	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			for _, in := range tx.Vin {
				newOutputs := TXOutputs{}
				key := append(utxoPrefix, in.Txid...)
				value, err := db.Get(key, nil)
				util.Handle(err)
				var outs TXOutputs
				outs.Deserialize(value)
				for outIdx, out := range outs.Outputs {
					if outIdx != in.Vout {
						newOutputs.Outputs = append(newOutputs.Outputs, out)
					}
				}
				if len(newOutputs.Outputs) == 0 {
					err := db.Delete(key, nil)
					util.Handle(err)
				} else {
					err := db.Put(key, newOutputs.Serialize(), nil)
					util.Handle(err)
				}
			}
		}
		newOutputs := TXOutputs{}
		newOutputs.Outputs = append(newOutputs.Outputs, tx.Vout...)
		key := append(utxoPrefix, tx.ID...)
		err := db.Put(key, newOutputs.Serialize(), nil)
		util.Handle(err)
	}
}

func (u *UTXOSet) DeleteByPrefix() {
	db := u.BlockChain.Database
	batch := new(leveldb.Batch)
	iter := db.NewIterator(nil, nil)
	defer iter.Release()
	for iter.Next() {
		if bytes.HasPrefix(iter.Key(), utxoPrefix) {
			fmt.Printf("Deleting %x\n", iter.Key())
			batch.Delete(iter.Key())
		}
		if batch.Len() >= max_number {
			db.Write(batch, nil)
			batch.Reset()
		}
	}
	db.Write(batch, nil)
}
