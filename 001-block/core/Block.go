package BLC

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"strconv"
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
}

func (block *Block) SetHash() {
	//1. Height []byte
	heightBytes := IntTOHex(block.Height)
	fmt.Println("height", heightBytes)

	//2. 将时间戳转 []byte 2到36
	timeString := strconv.FormatInt(block.Timestamp, 2)
	fmt.Println(timeString)
	timeBytes := []byte(timeString)
	fmt.Println(timeBytes)

	//3. 拼接所有属性
	blockBytes := bytes.Join([][]byte{heightBytes, block.PrevBlockHash, block.Data, timeBytes, block.Hash}, []byte{})

	//4. 生成Hash
	hash := sha256.Sum256(blockBytes)
	block.Hash = hash[:]
	fmt.Println("hash", block.Hash)
}

// 1. 创建新的区块
func NewBlock(data string, height int64, prevBlockHash []byte) *Block {

	block := &Block{height, prevBlockHash, []byte(data), time.Now().Unix(), nil}
	block.SetHash()
	return block
}

// 2. 生成创世区块
func CreateGenesisBlock(data string) *Block {

	return NewBlock(data, 1, make([]byte, 32))
}
