package node

import (
	"fmt"
	"learn_DumboNG/007-transaction/store"
	"os"
)

func (cli *CLI) printchain() {
	if store.DBExists() == false {
		fmt.Println("数据库不存在.....")
		os.Exit(1)
	}
	blockchain := store.BlockchainObject()
	defer blockchain.DB.Close()
	blockchain.Printchain()
}
