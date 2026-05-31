package node

import "learn_DumboNG/008-Wallet/store"

// 创建创世区块
func (cli *CLI) createGenesisBlockchain(address string) {
	blockchain := store.CreateBlockchainWithGenesisBlock(address)
	defer blockchain.DB.Close()

}
