package store

import (
	"fmt"
	"learn_DumboNG/006-Persistence_cli/core"
	"log"
	"math/big"
	"os"
	"time"

	"github.com/boltdb/bolt"
)

const dbName = "blockchain.db"
const blockTableName = "blocks"

type Blockchain struct {
	Tip []byte
	DB  *bolt.DB
}

// 6.返回区块链对象的方法
func BlockchainObject() *Blockchain {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			//读取最新区块的hash
			tip = b.Get([]byte("l"))
		}

		return nil
	})

	return &Blockchain{tip, db}
}

// 4.迭代器方法
func (blockchain *Blockchain) Iterator() *BlockchainIterator {

	return &BlockchainIterator{blockchain.Tip, blockchain.DB}
}

// 3.遍历输出所有区块的信息
func (blc *Blockchain) Printchain() {

	blockchainIterator := blc.Iterator()

	for {
		block := blockchainIterator.Next()

		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)

		fmt.Println()

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

}

// 2.增加区块到区块链里面
func (blc *Blockchain) AddBlocktoBlockchain(data string) {

	err := blc.DB.Update(func(tx *bolt.Tx) error {

		//1， 获取表
		b := tx.Bucket([]byte(blockTableName))

		//2. 创建新区块
		if b != nil {
			//获取最新区块
			blockBytes := b.Get(blc.Tip)
			block := core.DeserializeBlock(blockBytes)

			//3. 将区块序列化并存储到数据库
			newBlock := core.NewBlock(data, block.Height+1, block.Hash)
			err := b.Put(newBlock.Hash, newBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}

			//4. 更新数据库里面的“l"对应的hash
			err = b.Put([]byte("l"), newBlock.Hash)
			if err != nil {
				log.Panic(err)
			}

			//5. 更新blockchain的Tip
			blc.Tip = newBlock.Hash
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

// 1. 创建带有创世区块的区块链
// 判断数据库是否存在
func DBExists() bool {
	if _, err := os.Stat(dbName); os.IsNotExist(err) {
		return false
	}
	return true
}

func CreateBlockchainWithGenesisBlock(data string) {
	//判断数据库是否存在
	if DBExists() {
		fmt.Println("创世区块已存在......")
		os.Exit(1)
	}

	fmt.Println("正在创建创世区块...")

	//创建并打开数据库
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {
			log.Panic(err)
		}

		if b != nil {
			//创建创世区块
			genesisBlock := core.CreateGenesisBlock(data)
			//将创世区块存储到表中
			err := b.Put(genesisBlock.Hash, genesisBlock.Serialize())
			if err != nil {
				log.Panic(err)
			}
			//存储最新的区块的hash
			err = b.Put([]byte("l"), genesisBlock.Hash)
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
}
