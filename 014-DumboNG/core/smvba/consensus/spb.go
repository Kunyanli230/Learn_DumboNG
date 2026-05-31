package consensus

import (
	"learn_DumboNG/014-DumboNG/core"
	"learn_DumboNG/014-DumboNG/crypto"
	"learn_DumboNG/014-DumboNG/logger"
	"sync"
	"sync/atomic"
)

type SPB struct {
	c         *Core
	Proposer  core.NodeID
	Epoch     int64
	Round     int64
	BlockHash atomic.Value

	vm           sync.Mutex
	Votes        map[int8]int
	Voters       map[int8]map[core.NodeID]struct{}
	SentTwoPhase bool
	SentFinish   bool

	uvm              sync.Mutex
	unHandleVote     []*SPBVote
	unHandleProposal []*SPBProposal

	LockFlag atomic.Bool
}

func NewSPB(c *Core, epoch, round int64, proposer core.NodeID) *SPB {
	return &SPB{
		c:            c,
		Epoch:        epoch,
		Round:        round,
		Proposer:     proposer,
		unHandleVote: make([]*SPBVote, 0),
		Votes:        make(map[int8]int),
		Voters:       make(map[int8]map[core.NodeID]struct{}),
	}
}

func (s *SPB) processProposal(p *SPBProposal) {
	if p.Phase == SPB_ONE_PHASE {
		// already receive
		if p.B == nil || s.BlockHash.Load() != nil || p.Author != s.Proposer {
			return
		}
		blockHash := p.B.Hash()
		s.BlockHash.Store(blockHash)

		if vote, err := NewSPBVote(s.c.Name, p.Author, blockHash, s.Epoch, s.Round, p.Phase, s.c.SigService); err != nil {
			logger.Error.Printf("create spb vote message error:%v \n", err)
		} else {
			if s.c.Name != s.Proposer {
				s.c.Transimtor.Send(s.c.Name, s.Proposer, vote)
			} else {
				s.c.Transimtor.RecvChannel() <- vote
			}
		}

		s.uvm.Lock()
		for _, proposal := range s.unHandleProposal {
			go s.processProposal(proposal)
		}
		for _, vote := range s.unHandleVote {
			go s.processVote(vote)
		}
		s.unHandleProposal = nil
		s.unHandleVote = nil
		s.uvm.Unlock()

	} else if p.Phase == SPB_TWO_PHASE {
		if p.Author != s.Proposer {
			return
		}
		if s.BlockHash.Load() == nil {
			s.uvm.Lock()
			defer s.uvm.Unlock()
			s.unHandleProposal = append(s.unHandleProposal, p)
			return
		}
		//if lock ensure SPB_ONE_PHASE has received
		s.LockFlag.Store(true)
		if vote, err := NewSPBVote(s.c.Name, p.Author, crypto.Digest{}, s.Epoch, s.Round, p.Phase, s.c.SigService); err != nil {
			logger.Error.Printf("create spb vote message error:%v \n", err)
		} else {
			if s.c.Name != s.Proposer {
				s.c.Transimtor.Send(s.c.Name, s.Proposer, vote)
			} else {
				s.c.Transimtor.RecvChannel() <- vote
			}
		}
	}
}

func (s *SPB) processVote(p *SPBVote) {
	if p.Proposer != s.Proposer || p.Epoch != s.Epoch || p.Round != s.Round {
		return
	}
	blockHashValue := s.BlockHash.Load()
	if blockHashValue == nil {
		s.uvm.Lock()
		s.unHandleVote = append(s.unHandleVote, p)
		s.uvm.Unlock()
		return
	}
	blockHash := blockHashValue.(crypto.Digest)
	if p.Phase == SPB_ONE_PHASE && p.BlockHash != blockHash {
		return
	}
	if p.Phase == SPB_TWO_PHASE && p.BlockHash != (crypto.Digest{}) {
		return
	}

	s.vm.Lock()
	voters, ok := s.Voters[p.Phase]
	if !ok {
		voters = make(map[core.NodeID]struct{})
		s.Voters[p.Phase] = voters
	}
	if _, ok := voters[p.Author]; ok {
		s.vm.Unlock()
		return
	}
	voters[p.Author] = struct{}{}
	s.Votes[p.Phase]++
	num := s.Votes[p.Phase]
	shouldSendTwoPhase := num == s.c.Committee.HightThreshold() && p.Phase == SPB_ONE_PHASE && !s.SentTwoPhase
	shouldSendFinish := num == s.c.Committee.HightThreshold() && p.Phase == SPB_TWO_PHASE && !s.SentFinish
	if shouldSendTwoPhase {
		s.SentTwoPhase = true
	}
	if shouldSendFinish {
		s.SentFinish = true
	}
	s.vm.Unlock()
	// 2f+1 unique votes for the same phase/value.
	if shouldSendTwoPhase {
		if proposal, err := NewSPBProposal(
			s.c.Name,
			nil,
			s.Epoch,
			s.Round,
			SPB_TWO_PHASE,
			s.c.SigService,
		); err != nil {
			logger.Error.Printf("create spb proposal message error:%v \n", err)
		} else {
			s.c.Transimtor.Send(s.c.Name, core.NONE, proposal)
			s.c.Transimtor.RecvChannel() <- proposal
		}
	}
	if shouldSendFinish {
		if finish, err := NewFinish(s.c.Name, blockHash, s.Epoch, s.Round, s.c.SigService); err != nil {
			logger.Error.Printf("create finish message error:%v \n", err)
		} else {
			s.c.Transimtor.Send(s.c.Name, core.NONE, finish)
			s.c.Transimtor.RecvChannel() <- finish
		}
	}
}

func (s *SPB) IsLock() bool {
	return s.LockFlag.Load()
}

func (s *SPB) GetBlockHash() any {
	return s.BlockHash.Load()
}
