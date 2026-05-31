package crypto

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/big"
)

// 256位hash前面至少要有16个0
const targetBit = 16

type ProofOfWork struct {
	PrevBlockHash []byte
	TxHash        []byte
	Timestamp     int64
	Height        int64
	Hash          []byte
	target        *big.Int //不会溢出的存储
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.PrevBlockHash,
			pow.TxHash,
			IntTOHex(pow.Timestamp),
			IntTOHex(int64(targetBit)),
			IntTOHex(int64(nonce)),
			IntTOHex(pow.Height),
		},
		[]byte{},
	)

	return data
}

func (proofOfWork *ProofOfWork) IsValid() bool {
	//1.proofOfWork.Block.Hash 比较 proofOfWork.Target
	var hashInt big.Int
	hashInt.SetBytes(proofOfWork.Hash)
	if proofOfWork.target.Cmp(&hashInt) == 1 {
		return true
	}

	return false

}

func (proofOfWork *ProofOfWork) Run() ([]byte, int64) {
	//1.将Block的属性拼接成byte
	//2.生成hash
	//3.判断hash有效性（如果不满足条件就循环，满足就跳出循环）
	nonce := 0
	var hashInt big.Int // 存储新生成的hash
	var hash [32]byte
	for {
		//准备数据
		dataBytes := proofOfWork.prepareData(nonce)
		//生成hash
		hash = sha256.Sum256(dataBytes)
		fmt.Printf("\r%x\n", hash)
		//将hash存储到hashInt
		hashInt.SetBytes(hash[:])
		//判断hashInt是否小于Block里面的target
		if proofOfWork.target.Cmp(&hashInt) == 1 {
			break
		}
		nonce = nonce + 1
	}
	return hash[:], int64(nonce)
}

// 创建新的PoW对象
func NewProofOfWork(prevBlockHash []byte, txHash []byte, timestamp int64, height int64, hash []byte) *ProofOfWork {
	//1.创建一个初始值为1的target
	target := big.NewInt(1)
	//2.左移256减去targetBit
	target = target.Lsh(target, 256-targetBit)
	return &ProofOfWork{prevBlockHash, txHash, timestamp, height, hash, target}
}
