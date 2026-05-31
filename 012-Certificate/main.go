package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"learn_DumboNG/012-Certificate/certified"
	"os"
	"sort"
)

func main() {
	nodes := flag.Int("nodes", 4, "number of validators")
	faults := flag.Int("faults", 1, "number of silent faulty validators")
	rounds := flag.Int("rounds", 3, "number of data-plane rounds")
	batchSize := flag.Int("batch-size", 2, "transactions per block")
	jsonOut := flag.Bool("json", false, "print full simulation result as JSON")
	flag.Parse()

	result, err := certified.RunSimulation(*nodes, *faults, *rounds, *batchSize)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if *jsonOut {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	fmt.Printf("012 Certified Blocks data plane\n")
	fmt.Printf("nodes=%d faults=%d high_threshold=%d rounds=%d batch_size=%d\n", result.Committee.Size, result.Committee.Faults, result.Committee.HighThreshold(), *rounds, *batchSize)
	fmt.Printf("certificates=%d\n\n", len(result.Certificates))

	limit := len(result.Certificates)
	if limit > 12 {
		limit = 12
	}
	fmt.Println("Recent certificates:")
	for i := 0; i < limit; i++ {
		cert := result.Certificates[i]
		fmt.Printf("  proposer=%d height=%d hash=%s voters=%v\n", cert.Proposer, cert.Height, cert.BlockHash, cert.Voters)
	}
	if len(result.Certificates) > limit {
		fmt.Printf("  ... %d more\n", len(result.Certificates)-limit)
	}

	fmt.Println("\nCertified frontiers observed by honest nodes:")
	nodeIDs := make([]int, 0, len(result.Frontiers))
	for id := range result.Frontiers {
		nodeIDs = append(nodeIDs, int(id))
	}
	sort.Ints(nodeIDs)
	for _, nodeID := range nodeIDs {
		fmt.Printf("  node %d:", nodeID)
		frontier := result.Frontiers[certified.NodeID(nodeID)]
		proposers := make([]int, 0, len(frontier))
		for proposer := range frontier {
			proposers = append(proposers, int(proposer))
		}
		sort.Ints(proposers)
		for _, proposer := range proposers {
			cert := frontier[certified.NodeID(proposer)]
			fmt.Printf(" p%d->h%d", proposer, cert.Height)
		}
		fmt.Println()
	}
}
