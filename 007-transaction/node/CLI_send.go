package node

import (
	"fmt"
	"learn_DumboNG/007-transaction/store"
	"os"
)

// 转账
func (cli *CLI) send(from []string, to []string, amount []string) {
	if store.DBExists() == false {
		fmt.Println("数据库不存在.....")
		os.Exit(1)
	}
	blockchain := store.BlockchainObject()
	defer blockchain.DB.Close()

	blockchain.MineNewBlock(from, to, amount)
}
