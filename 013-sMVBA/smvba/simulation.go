package smvba

import "fmt"

type Event struct {
	Epoch  int    `json:"epoch"`
	Round  int    `json:"round"`
	Kind   string `json:"kind"`
	Node   NodeID `json:"node,omitempty"`
	Leader NodeID `json:"leader,omitempty"`
	Value  string `json:"value,omitempty"`
	Detail string `json:"detail,omitempty"`
}

type Result struct {
	Committee Committee         `json:"committee"`
	Epoch     int               `json:"epoch"`
	Decided   Value             `json:"decided"`
	Decisions map[NodeID]string `json:"decisions"`
	Events    []Event           `json:"events"`
}

type Simulator struct {
	Committee Committee
	Epoch     int
	MaxRounds int
}

func NewSimulator(nodes, faults, epoch, maxRounds int) (*Simulator, error) {
	committee, err := NewCommittee(nodes, faults)
	if err != nil {
		return nil, err
	}
	if maxRounds <= 0 {
		maxRounds = nodes + faults + 1
	}
	return &Simulator{Committee: committee, Epoch: epoch, MaxRounds: maxRounds}, nil
}

func (s *Simulator) Run() (*Result, error) {
	proposals := make(map[NodeID]Value, s.Committee.Size)
	locked := make(map[NodeID]Value, s.Committee.Size)
	result := &Result{
		Committee: s.Committee,
		Epoch:     s.Epoch,
		Decisions: make(map[NodeID]string),
		Events:    make([]Event, 0),
	}

	// Round-local SPB is modeled explicitly but compactly: every honest proposer
	// obtains 2f+1 SPB votes, then broadcasts Finish for its value.
	for _, proposer := range s.Committee.IDs() {
		if s.Committee.IsFaulty(proposer) {
			result.Events = append(result.Events, Event{Epoch: s.Epoch, Round: 0, Kind: "skip_faulty_spb", Node: proposer})
			continue
		}
		value := NewDemoValue(proposer, s.Epoch, s.Committee)
		proposals[proposer] = value
		locked[proposer] = value
		result.Events = append(result.Events,
			Event{Epoch: s.Epoch, Round: 0, Kind: "spb_proposal", Node: proposer, Value: value.Digest()},
			Event{Epoch: s.Epoch, Round: 0, Kind: "spb_vote_quorum", Node: proposer, Value: value.Digest(), Detail: fmt.Sprintf("%d votes", s.Committee.HighThreshold())},
			Event{Epoch: s.Epoch, Round: 0, Kind: "finish_quorum", Node: proposer, Value: value.Digest(), Detail: fmt.Sprintf("%d finish", s.Committee.HighThreshold())},
		)
	}

	for round := 0; round < s.MaxRounds; round++ {
		result.Events = append(result.Events, Event{Epoch: s.Epoch, Round: round, Kind: "done_quorum", Detail: fmt.Sprintf("%d done", s.Committee.HighThreshold())})
		leader := s.Leader(round)
		result.Events = append(result.Events, Event{Epoch: s.Epoch, Round: round, Kind: "threshold_coin", Leader: leader})

		value, ok := locked[leader]
		if !ok {
			result.Events = append(result.Events,
				Event{Epoch: s.Epoch, Round: round, Kind: "prevote_no", Leader: leader, Detail: "leader has no locked SPB value"},
				Event{Epoch: s.Epoch, Round: round, Kind: "finvote_no", Leader: leader, Detail: "advance to next round"},
			)
			continue
		}

		result.Events = append(result.Events,
			Event{Epoch: s.Epoch, Round: round, Kind: "prevote_yes", Leader: leader, Value: value.Digest(), Detail: fmt.Sprintf("%d yes votes", s.Committee.HighThreshold())},
			Event{Epoch: s.Epoch, Round: round, Kind: "finvote_commit", Leader: leader, Value: value.Digest(), Detail: fmt.Sprintf("%d commit votes", s.Committee.HighThreshold())},
			Event{Epoch: s.Epoch, Round: round, Kind: "halt", Leader: leader, Value: value.Digest()},
		)
		result.Decided = proposals[leader]
		for _, id := range s.Committee.HonestIDs() {
			result.Decisions[id] = value.Digest()
		}
		return result, nil
	}

	return nil, fmt.Errorf("sMVBA did not decide within %d rounds", s.MaxRounds)
}

// Leader stands in for the threshold-signature common coin used by 014. It is
// deterministic here so the teaching trace is reproducible.
func (s *Simulator) Leader(round int) NodeID {
	return NodeID((s.Epoch + round) % s.Committee.Size)
}
