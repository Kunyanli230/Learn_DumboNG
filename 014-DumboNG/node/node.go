package node

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"learn_DumboNG/014-DumboNG/config"
	"learn_DumboNG/014-DumboNG/core"
	smvba "learn_DumboNG/014-DumboNG/core/smvba/consensus"
	"learn_DumboNG/014-DumboNG/crypto"
	"learn_DumboNG/014-DumboNG/logger"
	"learn_DumboNG/014-DumboNG/pool"
	"learn_DumboNG/014-DumboNG/store"
	"net"
	"sync"
	"time"
)

type CommitInfo struct {
	Height    int64       `json:"height"`
	Proposer  core.NodeID `json:"proposer"`
	BatchID   int         `json:"batch_id"`
	TxCount   int         `json:"tx_count"`
	Timestamp int64       `json:"timestamp"`
}

type ControlRequest struct {
	Command string `json:"command"`
	Payload string `json:"payload,omitempty"`
}

type ControlResponse struct {
	OK      bool         `json:"ok"`
	Error   string       `json:"error,omitempty"`
	Status  any          `json:"status,omitempty"`
	Commits []CommitInfo `json:"commits,omitempty"`
	TxID    string       `json:"tx_id,omitempty"`
}

type Node struct {
	ID            int
	Committee     core.Committee
	TxPool        *pool.Pool
	commitChannel chan *smvba.Block

	mu          sync.Mutex
	commitCount int
	recent      []CommitInfo
}

func NewNode(
	keysFile, tssKeyFile, committeeFile, parametersFile, storePath, logPath string,
	logLevel, nodeID int,
) (*Node, error) {

	commitChannel := make(chan *smvba.Block, 1_000)
	//step 1: init log config
	logger.SetOutput(logger.InfoLevel, logger.NewFileWriter(fmt.Sprintf("%s/node-info-%d.log", logPath, nodeID)))
	logger.SetOutput(logger.DebugLevel, logger.NewFileWriter(fmt.Sprintf("%s/node-debug-%d.log", logPath, nodeID)))
	logger.SetOutput(logger.WarnLevel, logger.NewFileWriter(fmt.Sprintf("%s/node-warn-%d.log", logPath, nodeID)))
	logger.SetOutput(logger.ErrorLevel, logger.NewFileWriter(fmt.Sprintf("%s/node-error-%d.log", logPath, nodeID)))
	logger.SetLevel(logger.Level(logLevel))

	//step 2: ReadKeys
	_, priKey, err := config.GenKeysFromFile(keysFile)
	if err != nil {
		logger.Error.Println(err)
		return nil, err
	}

	shareKey, err := config.GenTsKeyFromFile(tssKeyFile)
	if err != nil {
		logger.Error.Println(err)
		return nil, err
	}

	//step 3: committee and parameters
	commitee, err := config.GenCommitteeFromFile(committeeFile)
	if err != nil {
		logger.Error.Println(err)
		return nil, err
	}

	poolParameters, coreParameters, err := config.GenParamatersFromFile(parametersFile)
	if err != nil {
		logger.Error.Println(err)
		return nil, err
	}

	//step 4: invoke pool and core
	txpool := pool.NewPool(poolParameters, commitee.Size(), nodeID)

	_store := store.NewStore(store.NewDefaultNutsDB(storePath))
	sigService := crypto.NewSigService(priKey, shareKey)

	n := &Node{
		ID:            nodeID,
		Committee:     commitee,
		TxPool:        txpool,
		commitChannel: commitChannel,
		recent:        make([]CommitInfo, 0, 64),
	}

	err = smvba.Consensus(core.NodeID(nodeID), commitee, coreParameters, txpool, _store, sigService, commitChannel)
	if err != nil {
		logger.Error.Println(err)
		return nil, err
	}
	logger.Info.Printf("Node %d successfully booted \n", nodeID)

	return n, nil
}

func (n *Node) Submit(payload []byte) string {
	digest := crypto.NewHasher().Add([]byte(fmt.Sprintf("%d", time.Now().UnixNano()))).Sum256(payload)
	n.TxPool.Submit(pool.Transaction(payload))
	return hex.EncodeToString(digest[:8])
}

func (n *Node) Status() map[string]any {
	n.mu.Lock()
	defer n.mu.Unlock()
	return map[string]any{
		"node_id":      n.ID,
		"committee":    n.Committee.Size(),
		"commit_count": n.commitCount,
		"recent_count": len(n.recent),
	}
}

func (n *Node) RecentCommits() []CommitInfo {
	n.mu.Lock()
	defer n.mu.Unlock()
	out := make([]CommitInfo, len(n.recent))
	copy(out, n.recent)
	return out
}

func (n *Node) StartControl(addr string) error {
	listen, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	logger.Info.Printf("Node %d control API listening on %s\n", n.ID, addr)
	go func() {
		for {
			conn, err := listen.Accept()
			if err != nil {
				logger.Warn.Printf("control accept error: %v\n", err)
				continue
			}
			go n.handleControl(conn)
		}
	}()
	return nil
}

func (n *Node) handleControl(conn net.Conn) {
	defer conn.Close()
	var req ControlRequest
	enc := json.NewEncoder(conn)
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		_ = enc.Encode(ControlResponse{OK: false, Error: err.Error()})
		return
	}

	switch req.Command {
	case "submit":
		txID := n.Submit([]byte(req.Payload))
		_ = enc.Encode(ControlResponse{OK: true, TxID: txID})
	case "status":
		_ = enc.Encode(ControlResponse{OK: true, Status: n.Status()})
	case "commits":
		_ = enc.Encode(ControlResponse{OK: true, Commits: n.RecentCommits()})
	case "peers":
		_ = enc.Encode(ControlResponse{OK: true, Status: n.Committee.Authorities})
	default:
		_ = enc.Encode(ControlResponse{OK: false, Error: "unknown command"})
	}
}

// AnalyzeBlock records committed blocks and blocks forever.
func (n *Node) AnalyzeBlock() {
	for block := range n.commitChannel {
		info := CommitInfo{
			Height:    block.Height,
			Proposer:  block.Proposer,
			BatchID:   block.Batch.ID,
			TxCount:   len(block.Batch.Txs),
			Timestamp: time.Now().UnixMilli(),
		}
		n.mu.Lock()
		n.commitCount++
		n.recent = append(n.recent, info)
		if len(n.recent) > 64 {
			n.recent = n.recent[len(n.recent)-64:]
		}
		n.mu.Unlock()
	}
}
