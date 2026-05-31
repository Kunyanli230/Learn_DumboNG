package core

type TXInput struct {
	//1.交易hash
	TxHash []byte
	//2.存储TXoutput的Vout的索引
	Vout int
	//3.用户名
	ScriptSig string
}

// 判断当前的消费是属于谁的钱
func (txInput *TXInput) UnLockWithAddress(address string) bool {
	return txInput.ScriptSig == address
}
