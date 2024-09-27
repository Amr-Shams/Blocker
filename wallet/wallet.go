package wallet

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

func NewPairKey() (ecdsa.PrivateKey, []byte) {
	Curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(Curve, rand.Reader)
	util.Handle(err)
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

func (w Wallet) MarshalJSON() ([]byte, error) {
	mapStringAny := map[string]any{
		"PrivateKey": map[string]any{
			"D": w.PrivateKey.D,
			"PublicKey": map[string]any{
				"X": w.PrivateKey.PublicKey.X,
				"Y": w.PrivateKey.PublicKey.Y,
			},
			"X": w.PrivateKey.X,
			"Y": w.PrivateKey.Y,
		},
		"PublicKey": w.PublicKey,
	}
	return json.Marshal(mapStringAny)
}
func (w *Wallet) UnmarshalJSON(data []byte) error {
	var mapStringAny map[string]interface{}
	err := json.Unmarshal(data, &mapStringAny)
	if err != nil {
		return err
	}

	privateKeyMap := mapStringAny["PrivateKey"].(map[string]interface{})
	publicKeyMap := privateKeyMap["PublicKey"].(map[string]interface{})

	w.PrivateKey.D = new(big.Int).SetInt64(int64(privateKeyMap["D"].(float64)))
	w.PrivateKey.PublicKey.Curve = elliptic.P256()
	w.PrivateKey.PublicKey.X = new(big.Int).SetInt64(int64(publicKeyMap["X"].(float64)))
	w.PrivateKey.PublicKey.Y = new(big.Int).SetInt64(int64(publicKeyMap["Y"].(float64)))

	w.PublicKey = []byte(mapStringAny["PublicKey"].(string))

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
func (ws *Wallets) NewTransaction(from, to string, amount int, UTXO *blockchain.UTXOSet, nodeId string) *blockchain.Transaction {
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
