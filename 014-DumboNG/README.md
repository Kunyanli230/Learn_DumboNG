# 014-DumboNG

当然014-DumboNG也并不是全方面超越ACS原语的，DumboNG的理论模型被称为unbounded memory BFT，在原论文当中的确是有一系列解决方法， 不过你也许已经发现我们这里实现的core协议状态没有 GC。

除此之外，这个项目还有大量需要优化的地方，甚至DumboNG原模型也在一直优化当中，不过我们这个项目先告一段落了！

## CLI用法

```bash
go run ./014-DumboNG --help
```

启动一个 4 节点本地测试网络，运行 10 秒：

```bash
go run ./014-DumboNG local --nodes 4 --duration 10s --rate 100 --batch-size 2
```

正常情况下，每个节点最后都会输出一个大于 0 的 `commit_count`，表示节点已经完成了区块提交。

示例输出片段：

```json
{
  "ok": true,
  "status": {
    "commit_count": 343,
    "committee": 4,
    "node_id": 0,
    "recent_count": 64
  }
}
```

## 生成运行目录

可以先生成运行所需的密钥、committee 和参数文件：

```bash
go run ./014-DumboNG init --nodes 4 --runtime 014-DumboNG/runtime
```

生成后的目录结构大致如下：

```text
014-DumboNG/runtime/
├── config/
│   ├── .committee.json
│   ├── .parameters.json
│   ├── .node-key-0.json
│   ├── .node-ts-key-0.json
│   └── ...
├── data/
└── logs/
```

其中：

- `.node-key-*.json` 是普通 ed25519 节点签名密钥；
- `.node-ts-key-*.json` 是阈值 BLS 签名份额；
- `.committee.json` 描述节点 ID、公钥和共识地址；
- `.parameters.json` 描述交易池和共识参数。

## 手动启动节点

生成 runtime 后，可以分别在多个终端里启动节点。

终端 1：

```bash
go run ./014-DumboNG node --id 0 --runtime 014-DumboNG/runtime
```

终端 2：

```bash
go run ./014-DumboNG node --id 1 --runtime 014-DumboNG/runtime
```

终端 3：

```bash
go run ./014-DumboNG node --id 2 --runtime 014-DumboNG/runtime
```

终端 4：

```bash
go run ./014-DumboNG node --id 3 --runtime 014-DumboNG/runtime
```

默认端口规则：

```text
共识端口：9000 + node_id
控制端口：10000 + node_id
```

例如：

```text
node 0 consensus = 127.0.0.1:9000
node 0 control   = 127.0.0.1:10000
node 1 consensus = 127.0.0.1:9001
node 1 control   = 127.0.0.1:10001
```

你也可以使用 `local` 命令启动多节点：

```bash
go run ./014-DumboNG local --nodes 4 --duration 20s --rate 1000 --batch-size 200
```

常用参数：

| 参数 | 含义 |
|---|---|
| `--nodes` | 节点数量 |
| `--duration` | 运行时长 |
| `--rate` | 每个节点模拟交易生成速率 |
| `--batch-size` | 每个 batch 包含的交易数 |
| `--runtime` | 运行目录 |
| `--base-port` | 共识端口起始值 |
| `--control-base-port` | 控制端口起始值 |
| `--faults` | 模拟故障节点数量 |
| `--log-level` | 日志等级 bitmask |

开启 debug 日志：

```bash
go run ./014-DumboNG local --nodes 4 --duration 10s --rate 100 --batch-size 2 --log-level 15
```

让本地节点持续运行，直到按 `Ctrl+C` 停止：

```bash
go run ./014-DumboNG local --nodes 4 --keep-running
```

## 与运行中的节点交互

节点启动后，可以通过 CLI 控制端口查询状态、提交交易、查看提交结果。

查询节点状态：

```bash
go run ./014-DumboNG status --id 0 --runtime 014-DumboNG/runtime
```

查看最近提交的区块：

```bash
go run ./014-DumboNG commits --id 0 --runtime 014-DumboNG/runtime
```

查看 committee 节点信息：

```bash
go run ./014-DumboNG peers --id 0 --runtime 014-DumboNG/runtime
```

提交一笔交易：

```bash
go run ./014-DumboNG submit --id 0 --runtime 014-DumboNG/runtime 'hello dumbo-ng'
```

## 交互式 console

也可以打开交互式控制台：

```bash
go run ./014-DumboNG console --id 0 --runtime 014-DumboNG/runtime
```

进入后可以输入：

```text
status
commits
peers
submit hello
quit
```

示例：

```text
Dumbo-NG console. Commands: status, commits, peers, submit TEXT, quit
dumbo-ng> status
dumbo-ng> submit hello
dumbo-ng> commits
dumbo-ng> quit
```

## 单独生成密钥

只生成普通节点密钥：

```bash
go run ./014-DumboNG keys --nodes 4 --runtime 014-DumboNG/runtime
```

只生成阈值签名密钥：

```bash
go run ./014-DumboNG threshold-keys --nodes 4 --threshold 3 --runtime 014-DumboNG/runtime
```

如果不指定 `--threshold`，默认使用 BFT 高阈值：

## 验证

普通测试：

```bash
go test ./014-DumboNG/...
```

Race 测试：

```bash
go test -race ./014-DumboNG/...
```

如果你在仓库根目录 `/home/kunyan/learn_consensus` 下执行，可以使用：

```bash
go -C learn_DumboNG test ./014-DumboNG/...
go -C learn_DumboNG test -race ./014-DumboNG/...
```

## 实现说明

当前实现包含以下核心模块：

```text
014-DumboNG/
├── config/   # 密钥、committee、参数文件生成和读取
├── core/     # 共识公共类型、Transmitor、SMVBA 核心
├── crypto/   # ed25519、BLS 阈值签名、hash 工具
├── logger/   # 日志输出
├── network/  # TCP Sender / Receiver / gob Codec
├── node/     # 节点组装、控制接口、commit 记录
├── pool/     # 交易池、模拟交易生成、手动交易提交
└── store/    # NutsDB 存储封装
```

共识流程大致为：

1. 节点从交易池中取 batch；
2. 生成普通 block；
3. 通过投票为 block 形成 certificate；
4. 当本地观察到足够多新的 block certificate 后进入 SMVBA；
5. 通过 SPB、Finish、Done、阈值签名公共随机数和 leader vote 完成一个 epoch 的决定；
6. 提交被 SMVBA 决定的 blocks；
7. 进入下一 epoch。
