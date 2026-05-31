package certified

import "fmt"

type voteKey struct {
	Proposer NodeID
	Height   int
	Hash     Digest
}

type voteCollector struct {
	voters    map[NodeID]struct{}
	certified bool
}

// Node models only the Dumbo-NG data plane. The first Faults node IDs are
// treated as silent faulty nodes by the simulator.
type Node struct {
	ID          NodeID
	Committee   Committee
	Faulty      bool
	nextHeight  int
	nextBatchID int
	lastHash    Digest

	Blocks      map[NodeID]map[int]Block
	CurrentCert map[NodeID]CertForBlockData
	collectors  map[voteKey]*voteCollector
}

func NewNode(id NodeID, committee Committee, faulty bool) *Node {
	current := make(map[NodeID]CertForBlockData, committee.Size)
	blocks := make(map[NodeID]map[int]Block, committee.Size)
	for _, nodeID := range committee.IDs() {
		current[nodeID] = CertForBlockData{Height: 0, Hash: ""}
		blocks[nodeID] = make(map[int]Block)
	}
	return &Node{
		ID:          id,
		Committee:   committee,
		Faulty:      faulty,
		nextHeight:  1,
		nextBatchID: int(id),
		Blocks:      blocks,
		CurrentCert: current,
		collectors:  make(map[voteKey]*voteCollector),
	}
}

func (n *Node) Propose(batchSize int) (BlockMessage, bool) {
	if n.Faulty {
		return BlockMessage{}, false
	}
	batch := Batch{ID: n.nextBatchID}
	for i := 0; i < batchSize; i++ {
		batch.Txs = append(batch.Txs, Tx(fmt.Sprintf("node-%d/tx-%d-%d", n.ID, n.nextHeight, i)))
	}
	block := Block{
		Proposer: n.ID,
		Height:   n.nextHeight,
		PrevHash: n.lastHash,
		Batch:    batch,
	}
	n.nextHeight++
	n.nextBatchID += n.Committee.Size
	n.lastHash = block.Hash()
	return BlockMessage{Author: n.ID, Block: block}, true
}

func (n *Node) HandleBlock(msg BlockMessage) (VoteForBlock, bool) {
	if n.Faulty {
		return VoteForBlock{}, false
	}
	if msg.Author != msg.Block.Proposer || msg.Block.Height <= 0 || msg.Block.Batch.Empty() {
		return VoteForBlock{}, false
	}
	if _, ok := n.Blocks[msg.Author]; !ok {
		n.Blocks[msg.Author] = make(map[int]Block)
	}
	if existing, ok := n.Blocks[msg.Author][msg.Block.Height]; ok && existing.Hash() != msg.Block.Hash() {
		return VoteForBlock{}, false
	}
	n.Blocks[msg.Author][msg.Block.Height] = msg.Block
	return VoteForBlock{
		Author:    n.ID,
		Proposer:  msg.Author,
		Height:    msg.Block.Height,
		BlockHash: msg.Block.Hash(),
	}, true
}

func (n *Node) HandleVote(vote VoteForBlock) (BlockCertificate, bool) {
	if n.Faulty || vote.Proposer != n.ID || vote.Height <= 0 {
		return BlockCertificate{}, false
	}
	key := voteKey{Proposer: vote.Proposer, Height: vote.Height, Hash: vote.BlockHash}
	collector, ok := n.collectors[key]
	if !ok {
		collector = &voteCollector{voters: make(map[NodeID]struct{})}
		n.collectors[key] = collector
	}
	if collector.certified {
		return BlockCertificate{}, false
	}
	collector.voters[vote.Author] = struct{}{}
	if len(collector.voters) < n.Committee.HighThreshold() {
		return BlockCertificate{}, false
	}
	collector.certified = true
	return newCertificate(vote.Proposer, vote.Height, vote.BlockHash, collector.voters), true
}

func (n *Node) HandleCertificate(cert BlockCertificate) {
	if n.Faulty {
		return
	}
	current := n.CurrentCert[cert.Proposer]
	if cert.Height > current.Height {
		n.CurrentCert[cert.Proposer] = certData(cert)
	}
}
