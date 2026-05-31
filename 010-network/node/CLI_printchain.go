package node

import "learn_DumboNG/010-network/store"

func (cli *CLI) printchain(nodeID string) {
	blockchain := store.BlockchainObject(nodeID)
	defer blockchain.DB.Close()
	blockchain.Printchain()
}
