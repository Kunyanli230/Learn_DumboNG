package certified

import "sort"

type Event struct {
	Round    int    `json:"round"`
	Kind     string `json:"kind"`
	Node     NodeID `json:"node"`
	Proposer NodeID `json:"proposer"`
	Height   int    `json:"height"`
	Hash     string `json:"hash,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

type Result struct {
	Committee    Committee                              `json:"committee"`
	Certificates []BlockCertificate                     `json:"certificates"`
	Frontiers    map[NodeID]map[NodeID]CertForBlockData `json:"frontiers"`
	Events       []Event                                `json:"events"`
}

func RunSimulation(nodes, faults, rounds, batchSize int) (*Result, error) {
	committee, err := NewCommittee(nodes, faults)
	if err != nil {
		return nil, err
	}
	if rounds <= 0 {
		rounds = 1
	}
	if batchSize <= 0 {
		batchSize = 1
	}

	validators := make(map[NodeID]*Node, nodes)
	for _, id := range committee.IDs() {
		validators[id] = NewNode(id, committee, int(id) < faults)
	}

	result := &Result{
		Committee:    committee,
		Certificates: make([]BlockCertificate, 0),
		Frontiers:    make(map[NodeID]map[NodeID]CertForBlockData),
		Events:       make([]Event, 0),
	}

	for round := 1; round <= rounds; round++ {
		for _, proposerID := range committee.IDs() {
			proposer := validators[proposerID]
			msg, ok := proposer.Propose(batchSize)
			if !ok {
				result.Events = append(result.Events, Event{Round: round, Kind: "skip_faulty_proposer", Node: proposerID, Proposer: proposerID})
				continue
			}
			blockHash := msg.Block.Hash()
			result.Events = append(result.Events, Event{Round: round, Kind: "broadcast_block", Node: proposerID, Proposer: proposerID, Height: msg.Block.Height, Hash: blockHash.Short()})

			for _, receiverID := range committee.IDs() {
				vote, voted := validators[receiverID].HandleBlock(msg)
				if !voted {
					continue
				}
				result.Events = append(result.Events, Event{Round: round, Kind: "vote_block", Node: receiverID, Proposer: proposerID, Height: vote.Height, Hash: vote.BlockHash.Short()})
				cert, certified := proposer.HandleVote(vote)
				if !certified {
					continue
				}
				result.Certificates = append(result.Certificates, cert)
				result.Events = append(result.Events, Event{Round: round, Kind: "certified", Node: proposerID, Proposer: proposerID, Height: cert.Height, Hash: cert.BlockHash})
				for _, nodeID := range committee.IDs() {
					validators[nodeID].HandleCertificate(cert)
				}
			}
		}
	}

	for _, id := range committee.IDs() {
		if validators[id].Faulty {
			continue
		}
		frontier := make(map[NodeID]CertForBlockData, committee.Size)
		for proposer, cert := range validators[id].CurrentCert {
			frontier[proposer] = cert
		}
		result.Frontiers[id] = frontier
	}

	sort.Slice(result.Certificates, func(i, j int) bool {
		if result.Certificates[i].Height != result.Certificates[j].Height {
			return result.Certificates[i].Height < result.Certificates[j].Height
		}
		return result.Certificates[i].Proposer < result.Certificates[j].Proposer
	})
	return result, nil
}
