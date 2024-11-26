package node

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"path/filepath"
	"text/template"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/storage"
	"github.com/Amr-Shams/Blocker/util"
	"github.com/Amr-Shams/Blocker/wallet"
)

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		// Handle preflight OPTIONS requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func handelPairRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	pairs := n.GetAllPeers()
	response, _ := json.Marshal(pairs)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handelInfoRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	fs, err := storage.GetFileSystem()
	if err != nil {
		http.Error(w, "Failed to get file system: "+err.Error(), http.StatusInternalServerError)
		return
	}
	data, err := fs.ReadFromFile()
	if err != nil {
		http.Error(w, "Failed to read from file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	n.SetTSXhistory(mapBytestoTSXhistory(data))
	response, err := json.Marshal(n)
	if err != nil {
		http.Error(w, "Failed to marshal response: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handelStatusRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	response, _ := json.Marshal(stateName[n.GetStatus()])
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handelBalanceRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	ws := n.GetWallets()
	var balances = make(map[string]int)
	UTXOSet := blockchain.UTXOSet{BlockChain: n.GetBlockchain()}
	for _, w := range ws.Wallets {
		pubKeyHash := util.Base58Decode([]byte(w.GetAddress()))
		pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-wallet.CheckSumLength]
		UTXOs := UTXOSet.FindUTXO(pubKeyHash)
		balance := 0
		for _, out := range UTXOs {
			fmt.Printf("Out: %v\n", out)
			balance += out.Value
		}
		fmt.Printf("Address: %s, Balance: %d\n", w.GetAddress(), balance)
		balances[w.GetAddress()] = balance
	}
	response, _ := json.Marshal(balances)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

// ADD(19): localStorage cache for post request 
// to handle the case of the user refreshing the page after sending a transaction
func handelSendRequest(w http.ResponseWriter, r *http.Request, n Node) {
	fromWallet, toWallet, amount, err := validateSendRequest(r)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	ws := n.GetWallets()
	if ws == nil || !ws.Authenticated(fromWallet) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
		return
	}
	fmt.Println("Creating transaction")
	UTXO := blockchain.UTXOSet{
		BlockChain: n.GetBlockchain(),
	}
	tx := ws.NewTransaction(fromWallet, toWallet, amount, &UTXO)
	if tx == nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Transaction failed"))
	}
	if fullNode, ok := n.(*FullNode); ok {
		fmt.Println("Adding transaction to mempool")
		err := fullNode.AddMemPool(tx)
		fmt.Println(err)
		if err == nil {
			fmt.Println("Broadcasting transaction")
			done := make(chan bool) // we communicate by sharing memory
			go func() {
				fullNode.BroadcastTSX(context.Background(), mapBlockchainTsxToServerTsx(tx))
				done <- true
			}()
			<-done
		} else {
			util.Handle(err)
		}
	} else {
		done := make(chan bool)
		go func() {
			if walletNode, ok := n.(*WalletNode); ok {
				walletNode.BroadcastTSX(context.Background(), mapBlockchainTsxToServerTsx(tx))
			}
			done <- true
		}()
		<-done
	}
	miningNode := SelectRandomFullNode(n)
	if miningNode != nil {
		conn, err := n.connectToPeer(miningNode.GetAddress())
		util.Handle(err)
		client := pb.NewBlockchainServiceClient(conn)
		res, err := client.MineBlock(context.Background(), &pb.Empty{})
		util.Handle(err)
		client.BroadcastBlock(context.Background(), res.Blocks[0])
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No mining node available"))
		return
	}
	tsxData := TSXhistory{
		From:   fromWallet,
		To:     toWallet,
		Amount: util.IntToStr(amount),
	}
	n.SetTSXhistory(append(n.GetTSXhistory(), &tsxData))
	fs, err := storage.GetFileSystem()
	util.Handle(err)
	dataBytes := mapTSXhistoryToBytes(&tsxData)
	err = fs.WriteToFile(dataBytes)
	util.Handle(err)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Transaction sent"))
}

func handelWalletsRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	ws := n.GetWallets()
	response, _ := json.Marshal(ws)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
func handelAddWalletRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	n.AddWallet()
	response, _ := json.Marshal(n.GetWallets())
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}

func handelMemPoolRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	if fullNode, ok := n.(*FullNode); ok {
		memPool := fullNode.GetMemPool()
		response, _ := json.Marshal(memPool)
		w.Header().Set("Content-Type", "application/json")
		w.Write(response)
	} else {
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("Method not allowed"))
	}
}

func handelIndexRequest(w http.ResponseWriter, _ *http.Request, n Node) {
	tmpl, err := template.ParseFiles(filepath.Join("public", "index.html"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	data := struct {
		Port string
	}{
		Port: n.GetAddress(),
	}
	tmpl.Execute(w, data)
}

func httpServer(l net.Listener, n Node) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handelIndexRequest(w, r, n)
	})
	mux.HandleFunc("/peers", func(w http.ResponseWriter, r *http.Request) {
		handelPairRequest(w, r, n)
	})
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		handelInfoRequest(w, r, n)
	})
	mux.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		handelStatusRequest(w, r, n)
	})
	mux.HandleFunc("/balance", func(w http.ResponseWriter, r *http.Request) {
		handelBalanceRequest(w, r, n)
	})
	mux.HandleFunc("/send", func(w http.ResponseWriter, r *http.Request) {
		handelSendRequest(w, r, n)
	})
	mux.HandleFunc("/wallets", func(w http.ResponseWriter, r *http.Request) {
		handelWalletsRequest(w, r, n)
	})
	mux.HandleFunc("/add-wallet", func(w http.ResponseWriter, r *http.Request) {
		handelAddWalletRequest(w, r, n)
	})
	mux.HandleFunc("/mempool", func(w http.ResponseWriter, r *http.Request) {
		handelMemPoolRequest(w, r, n)
	})
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("public"))))
	s := &http.Server{
		Handler: enableCORS(mux),
	}
	return s.Serve(l)
}
