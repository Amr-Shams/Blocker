package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/Amr-Shams/Blocker/util"
	"github.com/cbergoon/merkletree"
)

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

type TXInput struct {
	Txid      []byte
	Vout      int
	Signature []byte
	PubKey    []byte
}

type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

type TXOutputs struct {
	Outputs []TXOutput
}

func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := *NewOutTsx(100, to)
	tx := Transaction{nil, []TXInput{txin}, []TXOutput{txout}}
	tx.SetID()
	return &tx
}

func (tx *Transaction) SetID() {
	tx.ID, _ = tx.CalculateHash()
}

func (tx *Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := util.HashPubKey(in.PubKey)
	return bytes.Equal(lockingHash, pubKeyHash)
}

func (out *TXOutput) Lock(address string) {
	pubKeyHash := util.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-util.CheckSumLength]
	out.PubKeyHash = pubKeyHash
}

func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Equal(out.PubKeyHash, pubKeyHash)
}
func NewOutTsx(value int, address string) *TXOutput {
	tsx := TXOutput{value, nil}
	tsx.Lock(address)
	return &tsx
}
func (tx Transaction) String() string {
	var lines []string
	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Vin {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:      %x", input.Txid))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Vout))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}
	for i, output := range tx.Vout {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       PubKeyHash: %x", output.PubKeyHash))
	}
	return strings.Join(lines, "\n")
}
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput
	for _, vin := range tx.Vin {
		inputs = append(inputs, TXInput{vin.Txid, vin.Vout, nil, nil})
	}
	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOutput{vout.Value, vout.PubKeyHash})
	}
	return Transaction{tx.ID, inputs, outputs}
}
func ValidInput(tx *Transaction, prevTXs map[string]Transaction) error {
	for _, vin := range tx.Vin {
		if prevTXs[hex.EncodeToString(vin.Txid)].ID == nil {
			return errors.New("ERROR: Previous transaction is not correct")
		}
	}
	return nil
}
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}
	err := ValidInput(tx, prevTXs)
	util.Handle(err)
	txCopy := tx.TrimmedCopy()
	for inID, vin := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID, _ = txCopy.CalculateHash()
		txCopy.Vin[inID].PubKey = nil
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, txCopy.ID)
		util.Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)
		tx.Vin[inID].Signature = signature
	}
}
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true
	}
	err := ValidInput(tx, prevTXs)
	util.Handle(err)
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()
	for inID, vin := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.Txid)]
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID, _ = txCopy.CalculateHash()
		txCopy.Vin[inID].PubKey = nil
		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])
		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])
		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}
	return true
}
func (tx Transaction) CalculateHash() ([]byte, error) {
	var hash [32]byte
	txCopy := tx
	txCopy.ID = []byte{}
	hash = sha256.Sum256(txCopy.Serialize())
	return hash[:], nil
}

func (tx Transaction) Equals(other merkletree.Content) (bool, error) {
	otherT, ok := other.(Transaction)
	if !ok {
		return false, fmt.Errorf("could not assert other to Transaction")
	}

	return bytes.Equal(tx.ID, otherT.ID), nil
}

func (tx *Transaction) Serialize() []byte {
	return util.Serialize(tx)
}
func (tx *Transaction) Deserialize(data []byte) {
	util.Deserialize(data, tx)
}
func (outs *TXOutputs) Serialize() []byte {
	return util.Serialize(outs)
}
func (outs *TXOutputs) Deserialize(data []byte) {
	util.Deserialize(data, outs)
}
func (out *TXOutput) Serialize() []byte {
	return util.Serialize(out)
}
func (out *TXOutput) Deserialize(data []byte) {
	util.Deserialize(data, out)
}
