package BLC

type Blockchain struct {
	Blocks []*Block //存储有序的区块
}

// 增加区块到区块链里面
func (blc *Blockchain) AddBlocktoBlockchain(data string, height int64, preHash []byte) {
	//创造新区块
	newBlock := NewBlock(data, height, preHash)
	//往链里面添加新区块
	blc.Blocks = append(blc.Blocks, newBlock)
}

// 1. 创建带有创世区块的区块链
func CreateBlockchainWithGenesisBlock() *Blockchain {
	genesisBlock := CreateGenesisBlock("Genesis Data ...")
	return &Blockchain{[]*Block{genesisBlock}}
}
