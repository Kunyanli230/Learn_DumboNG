package certified

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
)

// NodeID identifies a validator in the teaching simulator.
type NodeID int

// Committee contains the static validator set and BFT thresholds.
type Committee struct {
	Size   int
	Faults int
}

func NewCommittee(size, faults int) (Committee, error) {
	if size <= 0 {
		return Committee{}, fmt.Errorf("nodes must be positive")
	}
	if faults < 0 {
		return Committee{}, fmt.Errorf("faults cannot be negative")
	}
	if size < 3*faults+1 {
		return Committee{}, fmt.Errorf("need n >= 3f+1, got n=%d f=%d", size, faults)
	}
	return Committee{Size: size, Faults: faults}, nil
}

// HighThreshold is the 2f+1 quorum used to certify a block.
func (c Committee) HighThreshold() int { return 2*c.Faults + 1 }

// LowThreshold is f+1, useful when explaining why at least one honest node is included.
func (c Committee) LowThreshold() int { return c.Faults + 1 }

func (c Committee) IDs() []NodeID {
	ids := make([]NodeID, c.Size)
	for i := range ids {
		ids[i] = NodeID(i)
	}
	return ids
}

// Tx is intentionally simple: 012 focuses on data availability and certificates,
// not transaction validity or UTXO execution.
type Tx string

type Batch struct {
	ID  int
	Txs []Tx
}

func (b Batch) Empty() bool { return b.ID < 0 || len(b.Txs) == 0 }

type Digest [32]byte

func (d Digest) Short() string { return hex.EncodeToString(d[:4]) }

// Block is the data-plane object that replaces the ACS/RBC batch payload in
// this chapter. A later sMVBA round will decide which certified block frontier
// should be committed.
type Block struct {
	Proposer NodeID
	Height   int
	PrevHash Digest
	Batch    Batch
}

func (b Block) Hash() Digest {
	h := sha256.New()
	writeInt(h, int64(b.Proposer))
	writeInt(h, int64(b.Height))
	h.Write(b.PrevHash[:])
	writeInt(h, int64(b.Batch.ID))
	for _, tx := range b.Batch.Txs {
		writeString(h, string(tx))
	}
	var d Digest
	copy(d[:], h.Sum(nil))
	return d
}

// CertForBlockData is the compact certificate frontier entry consumed by 013.
type CertForBlockData struct {
	Height int    `json:"height"`
	Hash   string `json:"hash"`
}

type BlockMessage struct {
	Author NodeID
	Block  Block
}

type VoteForBlock struct {
	Author    NodeID
	Proposer  NodeID
	Height    int
	BlockHash Digest
}

type BlockCertificate struct {
	Proposer  NodeID   `json:"proposer"`
	Height    int      `json:"height"`
	BlockHash string   `json:"block_hash"`
	Voters    []NodeID `json:"voters"`
}

func newCertificate(proposer NodeID, height int, hash Digest, voters map[NodeID]struct{}) BlockCertificate {
	ids := make([]NodeID, 0, len(voters))
	for id := range voters {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return BlockCertificate{
		Proposer:  proposer,
		Height:    height,
		BlockHash: hash.Short(),
		Voters:    ids,
	}
}

func certData(cert BlockCertificate) CertForBlockData {
	return CertForBlockData{Height: cert.Height, Hash: cert.BlockHash}
}

type hashWriter interface{ Write([]byte) (int, error) }

func writeInt(w hashWriter, v int64) {
	var buf [8]byte
	binary.BigEndian.PutUint64(buf[:], uint64(v))
	_, _ = w.Write(buf[:])
}

func writeString(w hashWriter, s string) {
	writeInt(w, int64(len(s)))
	_, _ = w.Write([]byte(s))
}
