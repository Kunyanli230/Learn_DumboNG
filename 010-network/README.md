Nakamoto共识的部分到这里就结束了！ 总结一下，你已经学到了：区块，出块，交易，地址，签名，通信，节点同步。

因为目的是高性能异步BFT，到这里我们就已经打好了学异步BFT基础，接下来我们可以开始搭建异步BFT了！

## 010-network 三节点 PowerShell 启动方式
 `$env:NODE_ID="端口号"; 命令`。

### 1. 进入项目根目录

### 2. 创建主节点钱包

```powershell
$env:NODE_ID="3000"; go run ./010-network createwallet
```

复制输出的 `Address`，下一步创建创世区块会用到。

### 3. 创建主节点创世区块

把 `<主节点钱包地址>` 替换成上一步输出的钱包地址：

```powershell
$env:NODE_ID="3000"; go run ../010-network createblockchain -address <主节点钱包地址>
```

### 4. 启动三个节点

现在需要打开三个 PowerShell 窗口，每个窗口都先进入项目根目录：

窗口 1：主节点 `localhost:3000`

```powershell
$env:NODE_ID="3000"; go run ./010-network startnode
```

窗口 2：钱包节点 `localhost:3001`

```powershell
$env:NODE_ID="3001"; go run ./010-network startnode
```

窗口 3：矿工节点 `localhost:3002`

```powershell
$env:NODE_ID="3002"; go run ./010-network startnode
```

如果要指定矿工奖励地址：

```powershell
$env:NODE_ID="3002"; go run ./010-network startnode -miner <矿工钱包地址>
```

### 5. 常用命令

查看当前节点钱包地址：

```powershell
$env:NODE_ID="3000"; go run ./010-network addresslists
```

查询余额：

```powershell
$env:NODE_ID="3000"; go run ./010-network getbalance -address <钱包地址>
```

打印当前节点区块链：

```powershell
$env:NODE_ID="3000"; go run ./010-network printchain
```

### 6. 重新开始/清理 011-network 本地数据

先停止所有节点，再执行：

```powershell
Remove-Item blockchain_3000.db, blockchain_3001.db, blockchain_3002.db -ErrorAction SilentlyContinue
Remove-Item 3000_wallets.dat, 3001_wallets.dat, 3002_wallets.dat -ErrorAction SilentlyContinue
Remove-Item blockchain_3000.db.lock, blockchain_3001.db.lock, blockchain_3002.db.lock -ErrorAction SilentlyContinue
```

### 7. 启动顺序说明
```
1. NODE_ID=3000 创建钱包
2. NODE_ID=3000 创建创世区块
3. 启动 3000 主节点
4. 启动 3001 钱包节点
5. 启动 3002 矿工节点
```

启动主节点 KnowNodes[localhost:3000]
-> 等待有人给主节点发送请求

启动钱包节点 [localhost:3001]
-> bc := GetObjectBlockchain(3001)
   sendVersion("localhost:3000", bc)
   ```[v,e,r,s,i,o,n,0,0,0,0,0] + Version{1, localhost:3001, 1}```
-> handleVersion()
-> 如果 bestBlockHeight > fromNodeHeight{
   -> 发送区块（3001，bc）
} 如果 bestBlockHeight < fromNodeHeight{
   -> 下载区块 {3001}
}
-> sendInv() + Inv{3000, "block", [][]byte}
-> handleInv()
-> sendGetBlocks() + GetBlocks{3000}
-> sendBlock
