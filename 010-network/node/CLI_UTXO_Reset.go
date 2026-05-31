package node

import "learn_DumboNG/010-network/store"

func (cli *CLI) resetUTXOSet(nodeID string) {
	blockchain := store.BlockchainObject(nodeID)
	defer blockchain.DB.Close()

	utxoSet := &store.UTXOSet{Blockchain: blockchain}
	utxoSet.ResetUTXOSet()
}
