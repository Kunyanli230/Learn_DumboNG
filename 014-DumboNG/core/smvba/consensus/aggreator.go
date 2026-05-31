package consensus

import (
	"learn_DumboNG/014-DumboNG/core"
	"learn_DumboNG/014-DumboNG/crypto"
	"sync"
)

type Aggreator struct {
	mu                 sync.Mutex
	committee          core.Committee
	finishAggreator    map[int64]map[int64]*FinishAggreator
	doneAggreator      map[int64]map[int64]*DoneAggreator
	prevoteAggreator   map[int64]map[int64]map[core.NodeID]*PreVoteAggreator
	finvoteAggreator   map[int64]map[int64]map[core.NodeID]*FinVoteAggreator
	haltAggreator      map[int64]map[int64]map[core.NodeID]map[crypto.Digest]*HaltAggreator
	blockvoteAggreator map[int64]map[crypto.Digest]*BlockVoteAggreator // map from height to block hash to votes
}

func NewAggreator(committee core.Committee) *Aggreator {
	return &Aggreator{
		committee:          committee,
		finishAggreator:    make(map[int64]map[int64]*FinishAggreator),
		doneAggreator:      make(map[int64]map[int64]*DoneAggreator),
		prevoteAggreator:   make(map[int64]map[int64]map[core.NodeID]*PreVoteAggreator),
		finvoteAggreator:   make(map[int64]map[int64]map[core.NodeID]*FinVoteAggreator),
		haltAggreator:      make(map[int64]map[int64]map[core.NodeID]map[crypto.Digest]*HaltAggreator),
		blockvoteAggreator: make(map[int64]map[crypto.Digest]*BlockVoteAggreator),
	}
}

func (a *Aggreator) AddFinishVote(finish *Finish) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	items, ok := a.finishAggreator[finish.Epoch]
	if !ok {
		items = make(map[int64]*FinishAggreator)
		a.finishAggreator[finish.Epoch] = items
	}
	item, ok := items[finish.Round]
	if !ok {
		item = NewFinishAggreator()
		items[finish.Round] = item
	}
	return item.Append(a.committee, finish)
}

func (a *Aggreator) AddDoneVote(done *Done) (int8, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	items, ok := a.doneAggreator[done.Epoch]
	if !ok {
		items = make(map[int64]*DoneAggreator)
		a.doneAggreator[done.Epoch] = items
	}
	item, ok := items[done.Round]
	if !ok {
		item = NewDoneAggreator()
		items[done.Round] = item
	}
	return item.Append(a.committee, done)
}

func (a *Aggreator) AddPreVote(vote *Prevote) (int8, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	rounds, ok := a.prevoteAggreator[vote.Epoch]
	if !ok {
		rounds = make(map[int64]map[core.NodeID]*PreVoteAggreator)
		a.prevoteAggreator[vote.Epoch] = rounds
	}
	leaders, ok := rounds[vote.Round]
	if !ok {
		leaders = make(map[core.NodeID]*PreVoteAggreator)
		rounds[vote.Round] = leaders
	}
	item, ok := leaders[vote.Leader]
	if !ok {
		item = NewPrevoteAggreator()
		leaders[vote.Leader] = item
	}
	return item.Append(a.committee, vote)
}

func (a *Aggreator) AddFinVote(vote *FinVote) (int8, crypto.Digest, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	rounds, ok := a.finvoteAggreator[vote.Epoch]
	if !ok {
		rounds = make(map[int64]map[core.NodeID]*FinVoteAggreator)
		a.finvoteAggreator[vote.Epoch] = rounds
	}
	leaders, ok := rounds[vote.Round]
	if !ok {
		leaders = make(map[core.NodeID]*FinVoteAggreator)
		rounds[vote.Round] = leaders
	}
	item, ok := leaders[vote.Leader]
	if !ok {
		item = NewFinVoteAggreator()
		leaders[vote.Leader] = item
	}
	return item.Append(a.committee, vote)
}

func (a *Aggreator) AddHaltVote(halt *Halt) (bool, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	rounds, ok := a.haltAggreator[halt.Epoch]
	if !ok {
		rounds = make(map[int64]map[core.NodeID]map[crypto.Digest]*HaltAggreator)
		a.haltAggreator[halt.Epoch] = rounds
	}
	leaders, ok := rounds[halt.Round]
	if !ok {
		leaders = make(map[core.NodeID]map[crypto.Digest]*HaltAggreator)
		rounds[halt.Round] = leaders
	}
	blocks, ok := leaders[halt.Leader]
	if !ok {
		blocks = make(map[crypto.Digest]*HaltAggreator)
		leaders[halt.Leader] = blocks
	}
	item, ok := blocks[halt.BlockHash]
	if !ok {
		item = NewHaltAggreator()
		blocks[halt.BlockHash] = item
	}
	return item.Append(a.committee, halt)
}

func (a *Aggreator) AddBlockVote(vote *VoteforBlock) (int8, error) {
	a.mu.Lock()
	defer a.mu.Unlock()

	items, ok := a.blockvoteAggreator[vote.Height]
	if !ok {
		items = make(map[crypto.Digest]*BlockVoteAggreator)
		a.blockvoteAggreator[vote.Height] = items
	}
	item, ok := items[vote.BlockHash]
	if !ok {
		item = NewBlockVoteAggreator()
		items[vote.BlockHash] = item
	}
	return item.Append(a.committee, vote)
}

type FinishAggreator struct {
	Authors map[core.NodeID]struct{}
}

func NewFinishAggreator() *FinishAggreator {
	return &FinishAggreator{
		Authors: make(map[core.NodeID]struct{}),
	}
}

func (f *FinishAggreator) Append(committee core.Committee, finish *Finish) (bool, error) {
	if _, ok := f.Authors[finish.Author]; ok {
		return false, core.ErrOneMoreMessage(finish.MsgType(), finish.Epoch, finish.Round, finish.Author)
	}
	f.Authors[finish.Author] = struct{}{}
	if len(f.Authors) == committee.HightThreshold() {
		return true, nil
	}
	return false, nil
}

const (
	DONE_LOW_FLAG int8 = iota
	DONE_HIGH_FLAG
	DONE_NONE_FLAG
)

type DoneAggreator struct {
	Authors map[core.NodeID]struct{}
}

func NewDoneAggreator() *DoneAggreator {
	return &DoneAggreator{
		Authors: make(map[core.NodeID]struct{}),
	}
}

func (d *DoneAggreator) Append(committee core.Committee, done *Done) (int8, error) {
	if _, ok := d.Authors[done.Author]; ok {
		return 0, core.ErrOneMoreMessage(done.MsgType(), done.Epoch, done.Round, done.Author)
	}
	d.Authors[done.Author] = struct{}{}
	if len(d.Authors) == committee.LowThreshold() {
		return DONE_LOW_FLAG, nil
	}
	if len(d.Authors) == committee.HightThreshold() {
		return DONE_HIGH_FLAG, nil
	}
	return DONE_NONE_FLAG, nil
}

const RANDOM_LEN = 3

type ElectAggreator struct {
	shares  []crypto.SignatureShare
	authors map[core.NodeID]struct{}
}

func NewElectAggreator() *ElectAggreator {
	return &ElectAggreator{
		shares:  make([]crypto.SignatureShare, 0),
		authors: make(map[core.NodeID]struct{}),
	}
}

func (e *ElectAggreator) Append(committee core.Committee, sigService *crypto.SigService, elect *ElectShare) (core.NodeID, error) {
	if _, ok := e.authors[elect.Author]; ok {
		return core.NONE, core.ErrOneMoreMessage(elect.MsgType(), elect.Epoch, elect.Round, elect.Author)
	}
	if err := crypto.VerifyTsPartial(sigService.ShareKey.PubPoly, elect.Hash(), elect.SigShare); err != nil {
		return core.NONE, err
	}
	e.authors[elect.Author] = struct{}{}
	e.shares = append(e.shares, elect.SigShare)
	if len(e.shares) == committee.HightThreshold() {
		coin, err := crypto.CombineIntactTSPartial(e.shares, sigService.ShareKey, elect.Hash())
		if err != nil {
			return core.NONE, err
		}
		var rand int
		for i := 0; i < RANDOM_LEN; i++ {
			if coin[i] > 0 {
				rand = rand<<8 + int(coin[i])
			} else {
				rand = rand<<8 + int(-coin[i])
			}
		}
		return core.NodeID(rand) % core.NodeID(committee.Size()), nil
	}
	return core.NONE, nil
}

const (
	ACTION_YES int8 = iota
	ACTION_NO
	ACTION_COMMIT
	ACTION_NONE
)

type PreVoteAggreator struct {
	authors map[core.NodeID]struct{}
	yesNums int64
	noNums  int64
	flag    bool
}

func NewPrevoteAggreator() *PreVoteAggreator {
	return &PreVoteAggreator{
		authors: make(map[core.NodeID]struct{}),
		yesNums: 0,
		noNums:  0,
		flag:    false,
	}
}

func (p *PreVoteAggreator) Append(committee core.Committee, vote *Prevote) (int8, error) {
	if _, ok := p.authors[vote.Author]; ok {
		return ACTION_NONE, core.ErrOneMoreMessage(vote.MsgType(), vote.Epoch, vote.Round, vote.Author)
	}
	p.authors[vote.Author] = struct{}{}
	if vote.Flag == VOTE_FLAG_NO {
		p.noNums++
	} else {
		p.yesNums++
	}

	if p.yesNums > 0 && !p.flag {
		p.flag = true
		return ACTION_YES, nil
	}
	if p.noNums == int64(committee.HightThreshold()) && !p.flag {
		p.flag = true
		return ACTION_NO, nil
	}
	return ACTION_NONE, nil
}

type FinVoteAggreator struct {
	authors   map[core.NodeID]struct{}
	yesByHash map[crypto.Digest]int64
	noNums    int64
}

func NewFinVoteAggreator() *FinVoteAggreator {
	return &FinVoteAggreator{
		authors:   make(map[core.NodeID]struct{}),
		yesByHash: make(map[crypto.Digest]int64),
		noNums:    0,
	}
}

func (f *FinVoteAggreator) Append(committee core.Committee, vote *FinVote) (int8, crypto.Digest, error) {
	if _, ok := f.authors[vote.Author]; ok {
		return ACTION_NONE, crypto.Digest{}, core.ErrOneMoreMessage(vote.MsgType(), vote.Epoch, vote.Round, vote.Author)
	}
	f.authors[vote.Author] = struct{}{}
	if vote.Flag == VOTE_FLAG_YES {
		f.yesByHash[vote.BlockHash]++
	} else {
		f.noNums++
	}

	th := int64(committee.HightThreshold())
	if f.yesByHash[vote.BlockHash] == th {
		return ACTION_COMMIT, vote.BlockHash, nil
	}
	if f.noNums == th {
		return ACTION_NO, crypto.Digest{}, nil
	}
	if int64(len(f.authors)) == th {
		var bestHash crypto.Digest
		var bestCount int64
		for hash, count := range f.yesByHash {
			if count > bestCount {
				bestHash = hash
				bestCount = count
			}
		}
		if bestCount > 0 {
			return ACTION_YES, bestHash, nil
		}
	}
	return ACTION_NONE, crypto.Digest{}, nil
}

type HaltAggreator struct {
	Authors map[core.NodeID]struct{}
}

func NewHaltAggreator() *HaltAggreator {
	return &HaltAggreator{
		Authors: make(map[core.NodeID]struct{}),
	}
}

func (h *HaltAggreator) Append(committee core.Committee, halt *Halt) (bool, error) {
	if _, ok := h.Authors[halt.Author]; ok {
		return false, core.ErrOneMoreMessage(halt.MsgType(), halt.Epoch, halt.Round, halt.Author)
	}
	h.Authors[halt.Author] = struct{}{}
	return len(h.Authors) == committee.HightThreshold(), nil
}

type BlockVoteAggreator struct {
	Authors map[core.NodeID]struct{}
}

func NewBlockVoteAggreator() *BlockVoteAggreator {
	return &BlockVoteAggreator{
		Authors: make(map[core.NodeID]struct{}),
	}
}

const (
	BV_LOW_FLAG int8 = iota
	BV_HIGH_FLAG
	BV_NONE_FLAG
)

func (b *BlockVoteAggreator) Append(committee core.Committee, vote *VoteforBlock) (int8, error) {
	if _, ok := b.Authors[vote.Author]; ok {
		return 0, core.ErrOneMoreMessage(vote.MsgType(), vote.Height, 0, vote.Author)
	}
	b.Authors[vote.Author] = struct{}{}
	if len(b.Authors) == committee.LowThreshold() {
		return BV_LOW_FLAG, nil
	}
	if len(b.Authors) == committee.HightThreshold() {
		return BV_HIGH_FLAG, nil
	}
	return BV_NONE_FLAG, nil
}
