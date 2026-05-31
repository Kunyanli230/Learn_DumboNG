package node

import "learn_DumboNG/008-Wallet/crypto"

func (cli *CLI) CreateWallet() {
	wallets, _ := crypto.NewWallets()
	wallets.CreateNewWallet()
}
