package node

import (
	"fmt"
	"learn_DumboNG/009-MerkleTree/crypto"
)

func (cli *CLI) addressLists() {
	fmt.Println("所有钱包地址如下:")
	wallets, _ := crypto.NewWallets()
	for address := range wallets.WalletsMap {
		fmt.Println(address)
	}
}
