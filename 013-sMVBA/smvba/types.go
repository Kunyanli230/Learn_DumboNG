package smvba

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"sort"
)

type NodeID int

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

func (c Committee) HighThreshold() int { return 2*c.Faults + 1 }
func (c Committee) LowThreshold() int  { return c.Faults + 1 }

func (c Committee) IDs() []NodeID {
	ids := make([]NodeID, c.Size)
	for i := range ids {
		ids[i] = NodeID(i)
	}
	return ids
}

func (c Committee) HonestIDs() []NodeID {
	ids := make([]NodeID, 0, c.Size-c.Faults)
	for _, id := range c.IDs() {
		if int(id) >= c.Faults {
			ids = append(ids, id)
		}
	}
	return ids
}

func (c Committee) IsFaulty(id NodeID) bool { return int(id) < c.Faults }

// CertForBlockData mirrors the output of 012 and the input consumed by 014.
type CertForBlockData struct {
	Height int    `json:"height"`
	Hash   string `json:"hash"`
}

// Value is the MVBA value: a candidate certified frontier.
type Value struct {
	Proposer NodeID                      `json:"proposer"`
	Epoch    int                         `json:"epoch"`
	Frontier map[NodeID]CertForBlockData `json:"frontier"`
}

func NewDemoValue(proposer NodeID, epoch int, committee Committee) Value {
	frontier := make(map[NodeID]CertForBlockData, committee.Size)
	for _, id := range committee.IDs() {
		height := 0
		if !committee.IsFaulty(id) {
			height = epoch + int(proposer) + int(id)
		}
		frontier[id] = CertForBlockData{Height: height, Hash: demoHash(proposer, id, epoch, height)}
	}
	return Value{Proposer: proposer, Epoch: epoch, Frontier: frontier}
}

func (v Value) Digest() string {
	h := sha256.New()
	writeInt(h, int64(v.Proposer))
	writeInt(h, int64(v.Epoch))
	ids := make([]int, 0, len(v.Frontier))
	for id := range v.Frontier {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, id := range ids {
		entry := v.Frontier[NodeID(id)]
		writeInt(h, int64(id))
		writeInt(h, int64(entry.Height))
		writeString(h, entry.Hash)
	}
	return hex.EncodeToString(h.Sum(nil)[:4])
}

func demoHash(proposer, node NodeID, epoch, height int) string {
	h := sha256.New()
	writeString(h, fmt.Sprintf("p%d-n%d-e%d-h%d", proposer, node, epoch, height))
	return hex.EncodeToString(h.Sum(nil)[:4])
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
