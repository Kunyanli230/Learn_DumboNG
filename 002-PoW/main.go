package main

import (
	"fmt"
	"learn_DumboNG/002-PoW/core"
)

func main() {
	////创世区块
	//blockchain := BLC.CreateBlockchainWithGenesisBlock()
	//
	////新区块
	//blockchain.AddBlocktoBlockchain("Send 100 USDT to KY", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	//blockchain.AddBlocktoBlockchain("Send 300 USDT to CZ", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	//blockchain.AddBlocktoBlockchain("Send 600 USDT to MM", blockchain.Blocks[len(blockchain.Blocks)-1].Height+1, blockchain.Blocks[len(blockchain.Blocks)-1].Hash)
	//fmt.Println(blockchain.Blocks)

	block := BLC.NewBlock("Test", 1, make([]byte, 32))
	fmt.Printf("%d\n", block.Nonce)
	fmt.Printf("%x\n", block.Hash)

	proofOfWork := BLC.NewProofOfWork(block)
	fmt.Printf("%v", proofOfWork.IsValid())
}
