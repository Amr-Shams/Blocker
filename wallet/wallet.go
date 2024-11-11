package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"math/big"

	"github.com/Amr-Shams/Blocker/blockchain"
	"github.com/Amr-Shams/Blocker/util"
)

const (
	CheckSumLength = 4
	Version        = byte(0x00)
)

type Wallet struct {
	PrivateKey ecdsa.PrivateKey
	PublicKey  []byte
}

func NewWallet() *Wallet {
	private, public := NewPairKey()
	return &Wallet{private, public}
}

// Function to generate a new key pair with debug statements
func NewPairKey() (ecdsa.PrivateKey, []byte) {
	Curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(Curve, rand.Reader)
	util.Handle(err)

	if private.D.Cmp(big.NewInt(0)) <= 0 || private.PublicKey.X.Cmp(big.NewInt(0)) <= 0 || private.PublicKey.Y.Cmp(big.NewInt(0)) <= 0 {
		log.Panic("Error: Invalid key pair")
	}
	pubKey := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)
	return *private, pubKey
}

func (w *Wallet) GetAddress() string {
	pubHash := util.HashPubKey(w.PublicKey)
	versionedPayload := append([]byte{Version}, pubHash...)
	checkSum := util.CheckSum(versionedPayload)
	fullPayload := append(versionedPayload, checkSum...)
	address := util.Base58Encode(fullPayload)
	return string(address)
}
func (w *Wallet) MarshalJSON() ([]byte, error) {
	mapStringAny := map[string]any{
		"PrivateKey": map[string]any{
			"D": w.PrivateKey.D.String(),
			"PublicKey": map[string]any{
				"X": w.PrivateKey.PublicKey.X.String(),
				"Y": w.PrivateKey.PublicKey.Y.String(),
			},
		},
		"PublicKey": hex.EncodeToString(w.PublicKey),
	}
	return json.Marshal(mapStringAny)
}

func (w *Wallet) UnmarshalJSON(data []byte) error {
	var mapStringAny map[string]interface{}
	if err := json.Unmarshal(data, &mapStringAny); err != nil {
		return err
	}

	privateKeyMap := mapStringAny["PrivateKey"].(map[string]interface{})
	publicKeyMap := privateKeyMap["PublicKey"].(map[string]interface{})
	w.PrivateKey = ecdsa.PrivateKey{}
	w.PrivateKey.PublicKey = ecdsa.PublicKey{}
	w.PrivateKey.D = new(big.Int)
	w.PrivateKey.PublicKey.X = new(big.Int)
	w.PrivateKey.PublicKey.Y = new(big.Int)
	w.PrivateKey.PublicKey.Curve = elliptic.P256()
	w.PublicKey = []byte{}
	d, ok := new(big.Int).SetString(privateKeyMap["D"].(string), 10)
	if !ok {
		err := errors.New("Error: Invalid D")
		util.Handle(err)
	}
	w.PrivateKey.D = d
	x, ok := new(big.Int).SetString(publicKeyMap["X"].(string), 10)
	if !ok {
		err := errors.New("Error: Invalid X")
		util.Handle(err)
	}
	w.PrivateKey.PublicKey.X = x
	y, ok := new(big.Int).SetString(publicKeyMap["Y"].(string), 10)
	if !ok {
		err := errors.New("Error: Invalid Y")
		util.Handle(err)
	}
	w.PrivateKey.PublicKey.Y = y

	w.PrivateKey.PublicKey.Curve = elliptic.P256()

	pubKeyStr, ok := mapStringAny["PublicKey"].(string)
	if !ok {
		err := errors.New("Error: Invalid PublicKey")
		util.Handle(err)
	}
	pubKeyBytes, err := hex.DecodeString(pubKeyStr)
	util.Handle(err)
	w.PublicKey = pubKeyBytes
	return nil
}
func (w Wallet) String() string {
	var str string
	str += "PrivateKey:\n"
	str += "D: " + w.PrivateKey.D.String() + "\n"
	str += "PublicKey:\n"
	str += "X: " + w.PrivateKey.PublicKey.X.String() + "\n"
	str += "Y: " + w.PrivateKey.PublicKey.Y.String() + "\n"
	str += "X: " + w.PrivateKey.X.String() + "\n"
	str += "Y: " + w.PrivateKey.Y.String() + "\n"
	str += "PublicKey: " + string(w.PublicKey) + "\n"
	return str
}
func (ws *Wallets) NewTransaction(from, to string, amount int, UTXO *blockchain.UTXOSet) *blockchain.Transaction {
	var inputs []blockchain.TXInput
	var outputs []blockchain.TXOutput
	w := ws.Wallets[from]
	pubKeyHash := util.Base58Decode([]byte(from))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-util.CheckSumLength]
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Panic("Error: Not enough funds")
	}
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		util.Handle(err)
		for _, out := range outs {
			input := blockchain.TXInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}
	outputs = append(outputs, *blockchain.NewOutTsx(amount, to))
	if acc > amount {
		outputs = append(outputs, *blockchain.NewOutTsx(acc-amount, from))
	}
	tx := blockchain.Transaction{nil, inputs, outputs}
	tx.SetID()
	UTXO.BlockChain.SignTransaction(&tx, w.PrivateKey)
	return &tx
}
