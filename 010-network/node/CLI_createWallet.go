package node

import (
	"fmt"
	"learn_DumboNG/010-network/crypto"
)

func (cli *CLI) CreateWallet(nodeID string) {
	wallets, _ := crypto.NewWallets(nodeID)
	wallets.CreateNewWallet(nodeID)
	fmt.Println(len(wallets.WalletsMap))
}
