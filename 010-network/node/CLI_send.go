package node

import (
	"fmt"
	"learn_DumboNG/010-network/store"
)

// 转账
func (cli *CLI) send(from []string, to []string, amount []string, nodeID string, mineNow bool) {
	blockchain := store.BlockchainObject(nodeID)
	defer blockchain.DB.Close()

	if mineNow {
		blockchain.MineNewBlock(from, to, amount, nodeID)
		utxoSet := &store.UTXOSet{Blockchain: blockchain}
		utxoSet.Update()
	} else {
		fmt.Println("由矿工节点处理...")
	}
}
