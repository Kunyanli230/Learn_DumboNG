package node

import "learn_DumboNG/009-MerkleTree/crypto"

func (cli *CLI) CreateWallet() {
	wallets, _ := crypto.NewWallets()
	wallets.CreateNewWallet()
}
