package node

import "learn_DumboNG/010-network/store"

func (cli *CLI) createGenesisBlockchain(address string, nodeID string) {
	blockchain := store.CreateBlockchainWithGenesisBlock(address, nodeID)
	defer blockchain.DB.Close()

	utxoSet := &store.UTXOSet{Blockchain: blockchain}
	utxoSet.ResetUTXOSet()
}
