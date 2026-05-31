package core

type TXOutput struct {
	Value        int64
	ScriptPubKey string //用户名
}

func (txOutput *TXOutput) UnLockScriptPubKeyWithAddress(address string) bool {
	return txOutput.ScriptPubKey == address
}
