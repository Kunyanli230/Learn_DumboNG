package consensus

import (
	"learn_DumboNG/014-DumboNG/core"
	"learn_DumboNG/014-DumboNG/crypto"
	"sync"
)

type Elector struct {
	mu              sync.Mutex
	leaders         map[int64]map[int64]core.NodeID
	electAggreators map[int64]map[int64]*ElectAggreator
	sigService      *crypto.SigService
	committee       core.Committee
}

func NewElector(sigService *crypto.SigService, committee core.Committee) *Elector {
	return &Elector{
		leaders:         make(map[int64]map[int64]core.NodeID),
		electAggreators: make(map[int64]map[int64]*ElectAggreator),
		sigService:      sigService,
		committee:       committee,
	}
}

func (e *Elector) SetLeader(epoch, round int64, leader core.NodeID) {
	e.mu.Lock()
	defer e.mu.Unlock()

	items, ok := e.leaders[epoch]
	if !ok {
		items = make(map[int64]core.NodeID)
		e.leaders[epoch] = items
	}
	items[round] = leader
}

func (e *Elector) Leader(epoch, round int64) core.NodeID {
	e.mu.Lock()
	defer e.mu.Unlock()

	items, ok := e.leaders[epoch]
	if !ok {
		items = make(map[int64]core.NodeID)
		e.leaders[epoch] = items
	}
	if node, ok := items[round]; ok {
		return node
	} else {
		return core.NONE
	}
}

func (e *Elector) AddShareVote(share *ElectShare) (core.NodeID, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	items, ok := e.electAggreators[share.Epoch]
	if !ok {
		items = make(map[int64]*ElectAggreator)
		e.electAggreators[share.Epoch] = items
	}

	eA, ok := items[share.Round]
	if !ok {
		eA = NewElectAggreator()
		items[share.Round] = eA
	}
	node, err := eA.Append(e.committee, e.sigService, share)
	if err != nil {
		return core.NONE, nil
	}
	if node != core.NONE {
		leaderItems, ok := e.leaders[share.Epoch]
		if !ok {
			leaderItems = make(map[int64]core.NodeID)
			e.leaders[share.Epoch] = leaderItems
		}
		leaderItems[share.Round] = node
	}
	return node, nil
}
