package consensus

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"learn_DumboNG/014-DumboNG/core"
	"learn_DumboNG/014-DumboNG/crypto"
	"learn_DumboNG/014-DumboNG/pool"
	"reflect"
	"sort"
)

const (
	SPB_ONE_PHASE int8 = iota
	SPB_TWO_PHASE
)

const (
	VOTE_FLAG_YES int8 = iota
	VOTE_FLAG_NO
)

type Validator interface {
	Verify(core.Committee) bool
}

type Block struct {
	Proposer core.NodeID
	Batch    pool.Batch
	//Epoch    int64
	Height  int64
	PreHash crypto.Digest
}

func NewBlock(proposer core.NodeID, Batch pool.Batch, Height int64, PreHash crypto.Digest) *Block {
	return &Block{
		Proposer: proposer,
		Batch:    Batch,
		Height:   Height,
		PreHash:  PreHash,
	}
}

func (b *Block) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *Block) Decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(b); err != nil {
		return err
	}
	return nil
}

func (b *Block) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(BlockMessageType))
	hashInt64(hasher, int64(b.Proposer))
	hashInt64(hasher, b.Height)
	hashDigest(hasher, b.PreHash)
	hashBatch(hasher, b.Batch)
	return hasher.Sum256(nil)
}

type BlockMessage struct {
	Author    core.NodeID
	B         *Block
	Height    int64
	Signature crypto.Signature
}

func NewBlockMessage(Author core.NodeID, B *Block, Height int64, sigService *crypto.SigService) (*BlockMessage, error) {
	blockMessage := &BlockMessage{
		Author: Author,
		B:      B,
		Height: Height,
	}
	sig, err := sigService.RequestSignature(blockMessage.Hash())
	if err != nil {
		return nil, err
	}
	blockMessage.Signature = sig
	return blockMessage, nil
}

func (bm *BlockMessage) Verify(committee core.Committee) bool {
	if bm.B == nil || bm.B.Proposer != bm.Author || bm.B.Height != bm.Height {
		return false
	}
	pub := committee.Name(bm.Author)
	return bm.Signature.Verify(pub, bm.Hash())
}

func (bm *BlockMessage) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(BlockMessageType))
	hashInt64(hasher, int64(bm.Author))
	hashInt64(hasher, bm.Height)
	if bm.B != nil {
		d := bm.B.Hash()
		hashDigest(hasher, d)
	}
	return hasher.Sum256(nil)
}

func (*BlockMessage) MsgType() int {
	return BlockMessageType
}

type VoteforBlock struct {
	Author    core.NodeID
	BlockHash crypto.Digest
	Height    int64
	Signature crypto.Signature
}

func NewVoteforBlock(Author core.NodeID, BlockHash crypto.Digest, Height int64, sigService *crypto.SigService) (*VoteforBlock, error) {
	vote := &VoteforBlock{
		Author:    Author,
		BlockHash: BlockHash,
		Height:    Height,
	}
	sig, err := sigService.RequestSignature(vote.Hash())
	if err != nil {
		return nil, err
	}
	vote.Signature = sig
	return vote, nil
}

func (v *VoteforBlock) Verify(committee core.Committee) bool {
	pub := committee.Name(v.Author)
	return v.Signature.Verify(pub, v.Hash())
}

func (v *VoteforBlock) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(VoteforBlockType))
	hashInt64(hasher, int64(v.Author))
	hashInt64(hasher, v.Height)
	hashDigest(hasher, v.BlockHash)
	return hasher.Sum256(nil)
}

func (*VoteforBlock) MsgType() int {
	return VoteforBlockType
}

type CertForBlockData struct {
	Height int64
	Hash   crypto.Digest
}

type SMVBABlock struct {
	Proposer      core.NodeID
	SMVBAProposal map[core.NodeID]*CertForBlockData
	Epoch         int64
}

func NewSMVBABlock(proposer core.NodeID, SMVBAProposal map[core.NodeID]*CertForBlockData, Epoch int64) *SMVBABlock {
	return &SMVBABlock{
		Proposer:      proposer,
		SMVBAProposal: SMVBAProposal,
		Epoch:         Epoch,
	}
}

func (b *SMVBABlock) Encode() ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(b); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (b *SMVBABlock) Decode(data []byte) error {
	buf := bytes.NewBuffer(data)
	if err := gob.NewDecoder(buf).Decode(b); err != nil {
		return err
	}
	return nil
}

func (b *SMVBABlock) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(SPBProposalType))
	hashInt64(hasher, int64(b.Proposer))
	hashInt64(hasher, b.Epoch)
	ids := make([]int, 0, len(b.SMVBAProposal))
	for id := range b.SMVBAProposal {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	hashInt64(hasher, int64(len(ids)))
	for _, id := range ids {
		cert := b.SMVBAProposal[core.NodeID(id)]
		hashInt64(hasher, int64(id))
		if cert == nil {
			hashInt64(hasher, -1)
			continue
		}
		hashInt64(hasher, cert.Height)
		hashDigest(hasher, cert.Hash)
	}
	return hasher.Sum256(nil)
}

type SPBProposal struct {
	Author    core.NodeID
	B         *SMVBABlock
	Epoch     int64
	Round     int64
	Phase     int8
	Signature crypto.Signature
}

func NewSPBProposal(Author core.NodeID, B *SMVBABlock, Epoch, Round int64, Phase int8, sigService *crypto.SigService) (*SPBProposal, error) {
	proposal := &SPBProposal{
		Author: Author,
		B:      B,
		Epoch:  Epoch,
		Round:  Round,
		Phase:  Phase,
	}
	sig, err := sigService.RequestSignature(proposal.Hash())
	if err != nil {
		return nil, err
	}
	proposal.Signature = sig
	return proposal, nil
}

func (p *SPBProposal) Verify(committee core.Committee) bool {
	if p.Phase == SPB_ONE_PHASE && p.B == nil {
		return false
	}
	if p.Phase == SPB_TWO_PHASE && p.B != nil {
		return false
	}
	pub := committee.Name(p.Author)
	return p.Signature.Verify(pub, p.Hash())
}

func (p *SPBProposal) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(SPBProposalType))
	hashInt64(hasher, int64(p.Author))
	hashInt64(hasher, p.Epoch)
	hashInt64(hasher, p.Round)
	hashInt8(hasher, p.Phase)
	if p.B != nil {
		d := p.B.Hash()
		hashDigest(hasher, d)
	}
	return hasher.Sum256(nil)
}

func (*SPBProposal) MsgType() int {
	return SPBProposalType
}

type SPBVote struct {
	Author    core.NodeID
	Proposer  core.NodeID
	BlockHash crypto.Digest
	Epoch     int64
	Round     int64
	Phase     int8
	Signature crypto.Signature
}

func NewSPBVote(Author, Proposer core.NodeID, BlockHash crypto.Digest, Epoch, Round int64, Phase int8, sigService *crypto.SigService) (*SPBVote, error) {
	vote := &SPBVote{
		Author:    Author,
		Proposer:  Proposer,
		BlockHash: BlockHash,
		Epoch:     Epoch,
		Round:     Round,
		Phase:     Phase,
	}
	sig, err := sigService.RequestSignature(vote.Hash())
	if err != nil {
		return nil, err
	}
	vote.Signature = sig
	return vote, nil
}

func (v *SPBVote) Verify(committee core.Committee) bool {
	pub := committee.Name(v.Author)
	return v.Signature.Verify(pub, v.Hash())
}

func (v *SPBVote) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(SPBVoteType))
	hashInt64(hasher, int64(v.Author))
	hashInt64(hasher, int64(v.Proposer))
	hashInt64(hasher, v.Epoch)
	hashInt64(hasher, v.Round)
	hashInt8(hasher, v.Phase)
	hashDigest(hasher, v.BlockHash)
	return hasher.Sum256(nil)
}

func (*SPBVote) MsgType() int {
	return SPBVoteType
}

type Finish struct {
	Author    core.NodeID
	BlockHash crypto.Digest
	Epoch     int64
	Round     int64
	Signature crypto.Signature
}

func NewFinish(Author core.NodeID, BlockHash crypto.Digest, Epoch, Round int64, sigService *crypto.SigService) (*Finish, error) {
	finish := &Finish{
		Author:    Author,
		BlockHash: BlockHash,
		Epoch:     Epoch,
		Round:     Round,
	}
	sig, err := sigService.RequestSignature(finish.Hash())
	if err != nil {
		return nil, err
	}
	finish.Signature = sig
	return finish, nil
}

func (f *Finish) Verify(committee core.Committee) bool {
	pub := committee.Name(f.Author)
	return f.Signature.Verify(pub, f.Hash())
}

func (f *Finish) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(FinishType))
	hashInt64(hasher, int64(f.Author))
	hashInt64(hasher, f.Epoch)
	hashInt64(hasher, f.Round)
	hashDigest(hasher, f.BlockHash)
	return hasher.Sum256(nil)
}

func (*Finish) MsgType() int {
	return FinishType
}

type Done struct {
	Author    core.NodeID
	Epoch     int64
	Round     int64
	Signature crypto.Signature
}

func NewDone(Author core.NodeID, epoch, round int64, sigService *crypto.SigService) (*Done, error) {
	done := &Done{
		Author: Author,
		Epoch:  epoch,
		Round:  round,
	}
	sig, err := sigService.RequestSignature(done.Hash())
	if err != nil {
		return nil, err
	}
	done.Signature = sig
	return done, nil
}

func (d *Done) Verify(committee core.Committee) bool {
	pub := committee.Name(d.Author)
	return d.Signature.Verify(pub, d.Hash())
}

func (d *Done) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(DoneType))
	hashInt64(hasher, int64(d.Author))
	hashInt64(hasher, d.Epoch)
	hashInt64(hasher, d.Round)
	return hasher.Sum256(nil)
}

func (*Done) MsgType() int {
	return DoneType
}

type ElectShare struct {
	Author   core.NodeID
	Epoch    int64
	Round    int64
	SigShare crypto.SignatureShare
}

func NewElectShare(Author core.NodeID, epoch, round int64, sigService *crypto.SigService) (*ElectShare, error) {
	elect := &ElectShare{
		Author: Author,
		Epoch:  epoch,
		Round:  round,
	}
	sig, err := sigService.RequestTsSugnature(elect.Hash())
	if err != nil {
		return nil, err
	}
	elect.SigShare = sig
	return elect, nil
}

func (e *ElectShare) Verify(committee core.Committee) bool {
	_ = committee.Name(e.Author)
	return e.SigShare.Verify(e.Hash())
}

func (e *ElectShare) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(ElectShareType))
	hashInt64(hasher, e.Epoch)
	hashInt64(hasher, e.Round)
	return hasher.Sum256(nil)
}

func (*ElectShare) MsgType() int {
	return ElectShareType
}

type Prevote struct {
	Author    core.NodeID
	Leader    core.NodeID
	Epoch     int64
	Round     int64
	Flag      int8
	BlockHash crypto.Digest
	Signature crypto.Signature
}

func NewPrevote(Author, Leader core.NodeID, Epoch, Round int64, flag int8, BlockHash crypto.Digest, sigService *crypto.SigService) (*Prevote, error) {
	prevote := &Prevote{
		Author:    Author,
		Leader:    Leader,
		Epoch:     Epoch,
		Round:     Round,
		Flag:      flag,
		BlockHash: BlockHash,
	}
	sig, err := sigService.RequestSignature(prevote.Hash())
	if err != nil {
		return nil, err
	}
	prevote.Signature = sig
	return prevote, nil
}

func (p *Prevote) Verify(committee core.Committee) bool {
	pub := committee.Name(p.Author)
	return p.Signature.Verify(pub, p.Hash())
}

func (p *Prevote) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(PrevoteType))
	hashInt64(hasher, int64(p.Author))
	hashInt64(hasher, int64(p.Leader))
	hashInt64(hasher, p.Epoch)
	hashInt64(hasher, p.Round)
	hashInt8(hasher, p.Flag)
	hashDigest(hasher, p.BlockHash)
	return hasher.Sum256(nil)
}

func (*Prevote) MsgType() int {
	return PrevoteType
}

type FinVote struct {
	Author    core.NodeID
	Leader    core.NodeID
	Epoch     int64
	Round     int64
	Flag      int8
	BlockHash crypto.Digest
	Signature crypto.Signature
}

func NewFinVote(Author, Leader core.NodeID, Epoch, Round int64, flag int8, BlockHash crypto.Digest, sigService *crypto.SigService) (*FinVote, error) {
	vote := &FinVote{
		Author:    Author,
		Leader:    Leader,
		Epoch:     Epoch,
		Round:     Round,
		Flag:      flag,
		BlockHash: BlockHash,
	}
	sig, err := sigService.RequestSignature(vote.Hash())
	if err != nil {
		return nil, err
	}
	vote.Signature = sig
	return vote, nil
}

func (p *FinVote) Verify(committee core.Committee) bool {
	pub := committee.Name(p.Author)
	return p.Signature.Verify(pub, p.Hash())
}

func (p *FinVote) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(FinVoteType))
	hashInt64(hasher, int64(p.Author))
	hashInt64(hasher, int64(p.Leader))
	hashInt64(hasher, p.Epoch)
	hashInt64(hasher, p.Round)
	hashInt8(hasher, p.Flag)
	hashDigest(hasher, p.BlockHash)
	return hasher.Sum256(nil)
}

func (*FinVote) MsgType() int {
	return FinVoteType
}

type Halt struct {
	Author    core.NodeID
	Epoch     int64
	Round     int64
	Leader    core.NodeID
	BlockHash crypto.Digest
	Signature crypto.Signature
}

func NewHalt(Author, Leader core.NodeID, BlockHash crypto.Digest, Epoch, Round int64, sigService *crypto.SigService) (*Halt, error) {
	h := &Halt{
		Author:    Author,
		Epoch:     Epoch,
		Round:     Round,
		Leader:    Leader,
		BlockHash: BlockHash,
	}
	sig, err := sigService.RequestSignature(h.Hash())
	if err != nil {
		return nil, err
	}
	h.Signature = sig
	return h, nil
}

func (h *Halt) Verify(committee core.Committee) bool {
	pub := committee.Name(h.Author)
	return h.Signature.Verify(pub, h.Hash())
}

func (h *Halt) Hash() crypto.Digest {
	hasher := crypto.NewHasher()
	hashInt64(hasher, int64(HaltType))
	hashInt64(hasher, int64(h.Author))
	hashInt64(hasher, h.Epoch)
	hashInt64(hasher, h.Round)
	hashInt64(hasher, int64(h.Leader))
	hashDigest(hasher, h.BlockHash)
	return hasher.Sum256(nil)
}

func (*Halt) MsgType() int {
	return HaltType
}

const (
	SPBProposalType int = iota
	SPBVoteType
	FinishType
	DoneType
	ElectShareType
	PrevoteType
	FinVoteType
	HaltType
	BlockMessageType
	VoteforBlockType
)

var DefaultMessageTypeMap = map[int]reflect.Type{
	SPBProposalType:  reflect.TypeOf(SPBProposal{}),
	SPBVoteType:      reflect.TypeOf(SPBVote{}),
	FinishType:       reflect.TypeOf(Finish{}),
	DoneType:         reflect.TypeOf(Done{}),
	ElectShareType:   reflect.TypeOf(ElectShare{}),
	PrevoteType:      reflect.TypeOf(Prevote{}),
	FinVoteType:      reflect.TypeOf(FinVote{}),
	HaltType:         reflect.TypeOf(Halt{}),
	BlockMessageType: reflect.TypeOf(BlockMessage{}),
	VoteforBlockType: reflect.TypeOf(VoteforBlock{}),
}

func hashInt64(hasher *crypto.Hasher, v int64) {
	var b [8]byte
	binary.BigEndian.PutUint64(b[:], uint64(v))
	hasher.Add(b[:])
}

func hashInt8(hasher *crypto.Hasher, v int8) {
	hasher.Add([]byte{byte(v)})
}

func hashDigest(hasher *crypto.Hasher, d crypto.Digest) {
	hasher.Add(d[:])
}

func hashBytes(hasher *crypto.Hasher, data []byte) {
	hashInt64(hasher, int64(len(data)))
	hasher.Add(data)
}

func hashBatch(hasher *crypto.Hasher, batch pool.Batch) {
	hashInt64(hasher, int64(batch.ID))
	hashInt64(hasher, int64(len(batch.Txs)))
	for _, tx := range batch.Txs {
		hashBytes(hasher, tx)
	}
}
