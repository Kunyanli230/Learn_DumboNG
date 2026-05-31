package main

import (
	"learn_DumboNG/004-Persistence/core"
)

func main() {
	//创世区块
	blockchain := BLC.CreateBlockchainWithGenesisBlock()
	defer blockchain.DB.Close()

	//新区块
	blockchain.AddBlocktoBlockchain("Send 100 USDT to KY")
	blockchain.AddBlocktoBlockchain("Send 300 USDT to CZ")
	blockchain.AddBlocktoBlockchain("Send 600 USDT to MM")

	blockchain.Printchain()
}
