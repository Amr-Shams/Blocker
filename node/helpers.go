package node

import (
	"encoding/json"
	"fmt"

	"github.com/Amr-Shams/Blocker/blockchain"
	pb "github.com/Amr-Shams/Blocker/server"
	"github.com/Amr-Shams/Blocker/wallet"
	"golang.org/x/exp/rand"
)

func MergeBlockChain(blockchain1, blockchain2 []*pb.Block) []*pb.Block {
	blockMap := make(map[string]*pb.Block)

	addToMap := func(blocks []*pb.Block) {
		for _, block := range blocks {
			blockHash := string(block.Hash)
			if existing, exists := blockMap[blockHash]; !exists ||
				(exists && shouldPreferBlock(block, existing)) {
				blockMap[blockHash] = block
			}
		}
	}

	addToMap(blockchain1)
	addToMap(blockchain2)
	var latestBlock *pb.Block
	for _, block := range blockMap {
		if latestBlock == nil || shouldPreferBlock(block, latestBlock) {
			latestBlock = block
		}
	}

	var mergedChain []*pb.Block
	currentBlock := latestBlock
	for currentBlock != nil && len(currentBlock.PrevHash) > 0 {
		mergedChain = append([]*pb.Block{currentBlock}, mergedChain...)
		currentBlock = blockMap[string(currentBlock.PrevHash)]
	}
	if currentBlock != nil && len(currentBlock.PrevHash) == 0 {
		mergedChain = append([]*pb.Block{currentBlock}, mergedChain...)
	}
	return mergedChain
}

func shouldPreferBlock(block1, block2 *pb.Block) bool {
	if block1.Timestamp != block2.Timestamp {
		return block1.Timestamp > block2.Timestamp
	}

	if block1.Nonce != block2.Nonce {
		return block1.Nonce > block2.Nonce
	}

	return string(block1.Hash) > string(block2.Hash)
}
func SelectRandomFullNode(n Node) *FullNode {
	peers := n.GetFullPeers()
	if len(peers) == 0 {
		return nil
	}
	idx := rand.Intn(len(peers))
	for _, peer := range peers {
		if idx == 0 {
			if peer.GetStatus() == StateIdle {
				return peer
			} else {
				idx = rand.Intn(len(peers))
			}
		}
		idx--
	}
	return nil
}

func mapServerWalletToBlockchainWallet(w *pb.Wallet) *wallet.Wallet {
	return &wallet.Wallet{
		PublicKey: w.PublicKey,
	}
}
func mapBlockchainWalletToServerWallet(wallet *wallet.Wallet) *pb.Wallet {
	if wallet == nil {
		return &pb.Wallet{}
	}
	return &pb.Wallet{
		PublicKey: wallet.PublicKey,
	}
}
func mapServerWalletsToBlockchainWallets(wallets *pb.Wallets) *wallet.Wallets {
	ws := &wallet.Wallets{Wallets: make(map[string]*wallet.Wallet)}
	for _, w := range wallets.Wallets {
		ws.Wallets[string(w.PublicKey)] = mapServerWalletToBlockchainWallet(w)
	}
	return ws
}
func mapBlockchainWalletsToServerWallets(wallets *wallet.Wallets) *pb.Wallets {
	ws := &pb.Wallets{}
	if wallets == nil || wallets.Wallets == nil {
		return ws
	}
	for _, w := range wallets.Wallets {
		ws.Wallets = append(ws.Wallets, mapBlockchainWalletToServerWallet(w))
	}
	return ws
}

func getStatusFromString(status string) int {
	switch status {
	case "Idle":
		return StateIdle
	case "Mining":
		return StateMining
	case "Connected":
		return StateConnected
	default:
		return StateIdle
	}
}
func SyncBlockchain(n Node, bestPeer *FullNode) error {
	blocks, err := n.FetchBlockchain(bestPeer)
	if blocks == nil {
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to fetch blockchain: %v", err)
	}
	if len(blocks) < int(n.GetBlockchain().GetHeight()) {
		return nil
	}
	if err := n.UpdateBlockchain(blocks); err != nil {
		return fmt.Errorf("failed to update blockchain: %v", err)
	}

	fmt.Printf("Chain of Node %s has %d blocks\n",
		n.GetAddress(), len(n.GetBlockchain().GetBlocks()))

	return nil
}
func mapServerTsxToBlockchainTsx(tsx *pb.Transaction) *blockchain.Transaction {
	Vin := []blockchain.TXInput{}
	for _, input := range tsx.Vin {
		Vin = append(Vin, blockchain.TXInput{
			Txid:      input.Txid,
			Vout:      int(input.Vout),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}
	Vout := []blockchain.TXOutput{}
	for _, output := range tsx.Vout {
		Vout = append(Vout, blockchain.TXOutput{
			Value:      int(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}
	return &blockchain.Transaction{
		ID:   tsx.Id,
		Vin:  Vin,
		Vout: Vout,
	}
}
func mapBlockchainTsxToServerTsx(tsx *blockchain.Transaction) *pb.Transaction {
	Vin := []*pb.TXInput{}
	for _, input := range tsx.Vin {
		Vin = append(Vin, &pb.TXInput{
			Txid:      input.Txid,
			Vout:      int32(input.Vout),
			Signature: input.Signature,
			PubKey:    input.PubKey,
		})
	}
	Vout := []*pb.TXOutput{}
	for _, output := range tsx.Vout {
		Vout = append(Vout, &pb.TXOutput{
			Value:      int32(output.Value),
			PubKeyHash: output.PubKeyHash,
		})
	}
	return &pb.Transaction{
		Id:   tsx.ID,
		Vin:  Vin,
		Vout: Vout,
	}
}
func mapClientBlocksToServerBlocks(blocks []*blockchain.Block) []*pb.Block {
	pbBlocks := []*pb.Block{}
	for _, block := range blocks {
		pbBlocks = append(pbBlocks, &pb.Block{
			Hash:         block.Hash,
			PrevHash:     block.PrevHash,
			Transactions: mapBlockchainTSXsToServerTSXs(block.Transactions),
			Nonce:        int32(block.Nonce),
			Timestamp:    block.TimeStamp.Unix(),
		})
	}
	return pbBlocks
}
func mapServerTSXsToBlockchainTSXs(tsxs []*pb.Transaction) []*blockchain.Transaction {
	res := []*blockchain.Transaction{}
	for _, tsx := range tsxs {
		res = append(res, mapServerTsxToBlockchainTsx(tsx))
	}
	return res
}
func mapBlockchainTSXsToServerTSXs(tsxs []*blockchain.Transaction) []*pb.Transaction {
	pbTSXs := []*pb.Transaction{}
	for _, tsx := range tsxs {
		pbTSXs = append(pbTSXs, mapBlockchainTsxToServerTsx(tsx))
	}
	return pbTSXs
}

func mapBytestoTSXhistory(data []byte) []*TSXhistory {
	history := []*TSXhistory{}
	err := json.Unmarshal(data, &history)
	if err == nil {
		return history
	}
	singleHistory := &TSXhistory{}
	err = json.Unmarshal(data, singleHistory)
	if err == nil {
		return []*TSXhistory{singleHistory}
	}
	return nil
}
func mapTSXhistoryToBytes(history *TSXhistory) []byte {
	data, err := json.Marshal(history)
	if err != nil {
		return nil
	}
	return data
}
