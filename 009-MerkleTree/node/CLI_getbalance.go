package node

import (
	"fmt"
	"learn_DumboNG/009-MerkleTree/store"
	"os"
)

// 查询余额
func (cli *CLI) getBalance(address string) {
	if store.DBExists() == false {
		fmt.Println("数据库不存在.....")
		os.Exit(1)
	}
	blockchain := store.BlockchainObject()
	defer blockchain.DB.Close()

	utxoSet := &store.UTXOSet{Blockchain: blockchain}
	amount := utxoSet.GetBalance(address)
	fmt.Printf("%s 的余额是: %d\n", address, amount)
}
