package core

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

// UTXOProvider 提供创建普通交易所需的 UTXO 查询能力。
type UTXOProvider interface {
	UnUTXOs(address string, txs []*Transaction) []*UTXO
	FindSpendableUTXOs(from string, amount int, txs []*Transaction) (int64, map[string][]int)
}

// UTXO交易模型
type Transaction struct {
	//1.交易hash ()
	TxHash []byte

	//2.输入
	Vins []*TXInput

	//3.输出
	Vouts []*TXOutput
}

// 判断当前交易是否是Coinbase交易
func (tx *Transaction) IsCoinbaseTransaction() bool {
	return len(tx.Vins) == 1 && len(tx.Vins[0].TxHash) == 0 && tx.Vins[0].Vout == -1
}

// transaction创建分两种情况
// 创世区块创建时的transaction
func NewCoinbaseTransaction(address string) *Transaction {

	txInput := &TXInput{[]byte{}, -1, "Genesis Block"}

	txOuptut := &TXOutput{10, address}

	txCoinbase := &Transaction{[]byte{}, []*TXInput{txInput}, []*TXOutput{txOuptut}}

	//设置hash值
	txCoinbase.HashTransaction()

	return txCoinbase
}

func (tx *Transaction) HashTransaction() {
	var result bytes.Buffer
	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(result.Bytes())
	tx.TxHash = hash[:]
}

// 转账时产生的transaction
func NewSimpleTransaction(from string, to string, amount int, blockchain UTXOProvider, txs []*Transaction) *Transaction {

	// 有一个函数，返回一个数组里面是from这个人所有未花费输出所对应的Transaction
	unUTXOs := blockchain.UnUTXOs(from, txs)
	fmt.Println(unUTXOs)
	// 有一个函数，返回一个总共消费的钱int和一个字典
	money, spendableUTXODic := blockchain.FindSpendableUTXOs(from, amount, txs)

	var txInputs []*TXInput
	var txOutputs []*TXOutput
	for txHash, indexArray := range spendableUTXODic {
		txHashBytes, _ := hex.DecodeString(string(txHash))
		//消费
		for _, index := range indexArray {
			txInput := &TXInput{txHashBytes, index, from}
			txInputs = append(txInputs, txInput)
		}
	}

	//转账
	txOutput := &TXOutput{int64(amount), to}
	txOutputs = append(txOutputs, txOutput)

	//找零
	txOutput = &TXOutput{int64(money) - int64(amount), from}
	txOutputs = append(txOutputs, txOutput)

	//设置hash
	tx := &Transaction{[]byte{}, txInputs, txOutputs}
	tx.HashTransaction()

	return tx
}
