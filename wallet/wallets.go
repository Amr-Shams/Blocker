package wallet

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Amr-Shams/Blocker/util"
)

var (
	StoreFile   = "./tmp/wallets_%s.data"
	MAX_WALLETS = 10
)

type Wallets struct {
	Wallets map[string]*Wallet
}

func CleanWallet(nodeId string) {
	StoreFile := fmt.Sprintf(StoreFile, nodeId)
	os.Remove(StoreFile)
}

func (ws *Wallets) Serialize() []byte {
	jsonData, err := json.Marshal(ws)
	util.Handle(err)
	return jsonData
}
func Deserialize(data []byte) *Wallets {
	var ws = Wallets{make(map[string]*Wallet)}
	err := json.Unmarshal(data, &ws)
	util.Handle(err)
	return &ws
}

func Save(ws *Wallets, nodeId string) {
	content := ws.Serialize()
	StoreFile := fmt.Sprintf(StoreFile, nodeId)
	if _, err := os.Stat(StoreFile); os.IsNotExist(err) {
		os.WriteFile(StoreFile, []byte("{}"), 0644)
	}
	err := os.WriteFile(StoreFile, content, 0644)
	util.Handle(err)
}
func LoadWallets(nodeId string) *Wallets {
	StoreFile := fmt.Sprintf(StoreFile, nodeId)
	content, err := os.ReadFile(StoreFile)
	if err != nil && os.IsNotExist(err) {
		return &Wallets{make(map[string]*Wallet)}
	}
	util.Handle(err)
	return Deserialize(content)
}
func CreateWallets(nodeId string) *Wallets {
	StoreFile := fmt.Sprintf(StoreFile, nodeId)
	if _, err := os.Stat(StoreFile); os.IsNotExist(err) {
		return &Wallets{make(map[string]*Wallet)}
	} else {
		return LoadWallets(nodeId)
	}
}
func (ws *Wallets) AddWallet(nodeId string) {
	if len(ws.Wallets) >= MAX_WALLETS {
		log.Panic("Max wallets reached")
	}
	w := NewWallet()
	address := w.GetAddress()
	ws.Wallets[address] = w
	Save(ws, nodeId)
}

func (ws *Wallets) GetWallet(address string) *Wallet {
	return ws.Wallets[address]
}
func (ws *Wallets) Authenticated(address string) bool {
	_, ok := ws.Wallets[address]
	return ok
}
func GetAllAddresses(nodeId string) {
	Wallets := LoadWallets(nodeId).Wallets
	for address := range Wallets {
		fmt.Println(address)
	}
}
