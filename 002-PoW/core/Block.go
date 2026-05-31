package BLC

import (
	"time"
)

type Block struct {
	//1. 区块高度
	Height int64
	//2. 上一个区块hash
	PrevBlockHash []byte
	//3. 交易数据
	Data []byte
	//4. 时间戳
	Timestamp int64
	//5。 当前区块hash
	Hash []byte
	//6. Nonce
	Nonce int64
}

// 1. 创建新的区块
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {
	//创建区块
	block := &Block{height, prevBlockHash, []byte(data), time.Now().Unix(), nil, 0}
	//调用PoW并且返回有效hash和nonce
	pow := NewProofOfWork(block)
	// 挖矿验证
	hash, nonce := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// 2. 生成创世区块
func CreateGenesisBlock(data string) *Block {

	return NewBlock(data, 1, make([]byte, 32))
}
