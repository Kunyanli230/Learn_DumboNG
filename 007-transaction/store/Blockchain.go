package store

import (
	"encoding/hex"
	"fmt"
	"learn_DumboNG/007-transaction/core"
	"log"
	"math/big"
	"os"
	"strconv"
	"time"

	"github.com/boltdb/bolt"
)

const dbName = "blockchain.db"
const blockTableName = "blocks"

type Blockchain struct {
	Tip []byte
	DB  *bolt.DB
}

// 转账时查找可用的UTXO
func (blockchain *Blockchain) FindSpendableUTXOs(from string, amount int, txs []*core.Transaction) (int64, map[string][]int) {

	//1.获取所有的UTXO
	utxos := blockchain.UnUTXOs(from, txs)

	//2.遍历所有的UTXO，找到可用的UTXO
	spendableUTXO := make(map[string][]int)
	var value int64
	for _, utxo := range utxos {
		value = value + utxo.Output.Value

		hash := hex.EncodeToString(utxo.TxHash)
		spendableUTXO[hash] = append(spendableUTXO[hash], utxo.Index)

		if value >= int64(amount) {
			break
		}
	}
	if value < int64(amount) {
		fmt.Println("&s 的余额不足", from)
		os.Exit(1)
	}
	return value, spendableUTXO
}

// 9.查询余额
func (blockchain *Blockchain) GetBalance(address string) int64 {
	utxos := blockchain.UnUTXOs(address, []*core.Transaction{})
	var amount int64
	for _, utxo := range utxos {
		amount += utxo.Output.Value
	}
	return amount
}

// 8.如果一个地址对应的TXOutput未花费，添加到数组中并返回该地址对应的所有未花费的Transaction
func (blockchain *Blockchain) UnUTXOs(address string, txs []*core.Transaction) []*core.UTXO {

	var unUTXOs []*core.UTXO
	spentTXOutputs := make(map[string][]int) //我要一个字典 {hash: [index1, index2, ...]}

	for _, tx := range txs {
		if tx.IsCoinbaseTransaction() == false {
			for _, in := range tx.Vins {
				if in.UnLockWithAddress(address) {
					key := hex.EncodeToString(in.TxHash)
					spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
				}
			}
		}
	}

	for _, tx := range txs {
	Work1:
		for index, out := range tx.Vouts {
			if out.UnLockScriptPubKeyWithAddress(address) {
				if len(spentTXOutputs) == 0 {
					utxo := &core.UTXO{
						TxHash: tx.TxHash,
						Index:  index,
						Output: out,
					}
					unUTXOs = append(unUTXOs, utxo)
				} else {
					for hash, indexArray := range spentTXOutputs {
						txHashStr := hex.EncodeToString(tx.TxHash)
						if hash == txHashStr {
							var isUnspentUTXO bool
							for _, outIndex := range indexArray {
								if index == outIndex {
									isUnspentUTXO = true
									continue Work1
								}
								if isUnspentUTXO == false {
									utxo := &core.UTXO{
										TxHash: tx.TxHash,
										Index:  index,
										Output: out,
									}
									unUTXOs = append(unUTXOs, utxo)
								}
							}

						} else {
							utxo := &core.UTXO{
								TxHash: tx.TxHash,
								Index:  index,
								Output: out,
							}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}

		}
	}

	blockIterator := blockchain.Iterator()

	for {
		block := blockIterator.Next()
		fmt.Println(block)
		fmt.Println()

		for i := len(block.Txs) - 1; i >= 0; i-- {
			tx := block.Txs[i]
			// txHash

			//Vins
			if tx.IsCoinbaseTransaction() == false {
				for _, in := range tx.Vins {
					if in.UnLockWithAddress(address) {
						key := hex.EncodeToString(in.TxHash)
						spentTXOutputs[key] = append(spentTXOutputs[key], in.Vout)
					}
				}
			}
			//Vouts
		work:
			for index, out := range tx.Vouts {
				if out.UnLockScriptPubKeyWithAddress(address) {
					if spentTXOutputs != nil {
						if len(spentTXOutputs) != 0 {
							var isSpentUTXO bool
							for txHash, indexArray := range spentTXOutputs {
								for _, i := range indexArray {
									if index == i && txHash == hex.EncodeToString(tx.TxHash) {
										isSpentUTXO = true
										continue work
									}
								}
							}
							if isSpentUTXO == false {
								utxo := &core.UTXO{
									TxHash: tx.TxHash,
									Index:  index,
									Output: out,
								}
								unUTXOs = append(unUTXOs, utxo)
							}
						} else {
							utxo := &core.UTXO{
								TxHash: tx.TxHash,
								Index:  index,
								Output: out,
							}
							unUTXOs = append(unUTXOs, utxo)
						}
					}
				}
			}

		}

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if hashInt.Cmp(big.NewInt(0)) == 0 {
			break
		}
	}

	return unUTXOs
}

// 7.挖掘新的区块
func (blockchain *Blockchain) MineNewBlock(from []string, to []string, amount []string) {

	// 建立一笔交易
	fmt.Println(from)
	fmt.Println(to)
	fmt.Println(amount)

	var txs []*core.Transaction
	for index, address := range from {
		value, _ := strconv.Atoi(amount[index])
		tx := core.NewSimpleTransaction(address, to[index], value, blockchain, txs)
		txs = append(txs, tx)
		fmt.Println(tx)
	}

	// 建立交易数组
	var block *core.Block

	blockchain.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			hash := b.Get([]byte("l"))
			blockBytes := b.Get(hash)

			block = core.DeserializeBlock(blockBytes)
		}
		return nil
	})

	// 建立新的区块
	block = core.NewBlock(txs, block.Height+1, block.Hash)

	// 更新区块链的Tip
	blockchain.DB.Update(func(tx *bolt.Tx) error {

		b := tx.Bucket([]byte(blockTableName))
		if b != nil {
			b.Put(block.Hash, block.Serialize())
			b.Put([]byte("l"), block.Hash)
			blockchain.Tip = block.Hash
		}
		return nil
	})

}

// 6.返回区块链对象的方法
func BlockchainObject() *Blockchain {
	db, err := bolt.Open(dbName, 0600, nil)
	if err != nil {
		log.Fatal(err)
	}

	var tip []byte

	err = db.View(func(tx *bolt.Tx) error {

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
		fmt.Printf("Timestamp: %s\n", time.Unix(block.Timestamp, 0).Format("2006-01-02 03:04:05 PM"))
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Nonce: %d\n", block.Nonce)
		fmt.Println("Txs :")
		for _, tx := range block.Txs {
			fmt.Printf("%x\n", tx.TxHash)
			fmt.Println("Vins :")
			for _, in := range tx.Vins {
				fmt.Printf("%x\n", in.TxHash)
				fmt.Printf("%d\n", in.Vout)
				fmt.Printf("%s\n", in.ScriptSig)
			}

			fmt.Println("Vouts:")
			for _, out := range tx.Vouts {
				fmt.Println(out.Value)
				fmt.Println(out.ScriptPubKey)
			}

		}

		fmt.Println()

		var hashInt big.Int
		hashInt.SetBytes(block.PrevBlockHash)
		if big.NewInt(0).Cmp(&hashInt) == 0 {
			break
		}
	}

}

// 2.增加区块到区块链里面
func (blc *Blockchain) AddBlocktoBlockchain(txs []*core.Transaction) {

	err := blc.DB.Update(func(tx *bolt.Tx) error {

		//1， 获取表
		b := tx.Bucket([]byte(blockTableName))

		//2. 创建新区块
		if b != nil {
			//获取最新区块
			blockBytes := b.Get(blc.Tip)
			block := core.DeserializeBlock(blockBytes)

			//3. 将区块序列化并存储到数据库
			newBlock := core.NewBlock(txs, block.Height+1, block.Hash)
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

func CreateBlockchainWithGenesisBlock(address string) *Blockchain {
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
	var genesisHash []byte
	err = db.Update(func(tx *bolt.Tx) error {

		b, err := tx.CreateBucket([]byte(blockTableName))
		if err != nil {
			log.Panic(err)
		}

		if b != nil {
			txCoinbase := core.NewCoinbaseTransaction(address)
			//创建创世区块
			genesisBlock := core.CreateGenesisBlock([]*core.Transaction{txCoinbase})
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

			genesisHash = genesisBlock.Hash
		}

		return nil
	})

	return &Blockchain{Tip: genesisHash, DB: db}
}
