package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"learn_DumboNG/013-sMVBA/smvba"
	"os"
	"sort"
)

func main() {
	nodes := flag.Int("nodes", 4, "number of validators")
	faults := flag.Int("faults", 1, "number of silent faulty validators")
	epoch := flag.Int("epoch", 0, "sMVBA epoch")
	maxRounds := flag.Int("max-rounds", 0, "maximum MVBA rounds; 0 chooses a safe default")
	jsonOut := flag.Bool("json", false, "print full simulation result as JSON")
	flag.Parse()

	sim, err := smvba.NewSimulator(*nodes, *faults, *epoch, *maxRounds)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	result, err := sim.Run()
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

	fmt.Println("013 sMVBA control plane")
	fmt.Printf("nodes=%d faults=%d high_threshold=%d epoch=%d\n", result.Committee.Size, result.Committee.Faults, result.Committee.HighThreshold(), result.Epoch)
	fmt.Printf("decided proposer=%d value=%s\n\n", result.Decided.Proposer, result.Decided.Digest())

	fmt.Println("Trace:")
	for _, event := range result.Events {
		fmt.Printf("  e%d/r%d %-17s", event.Epoch, event.Round, event.Kind)
		if event.Node != 0 || event.Kind == "skip_faulty_spb" || event.Kind == "spb_proposal" || event.Kind == "spb_vote_quorum" || event.Kind == "finish_quorum" {
			fmt.Printf(" node=%d", event.Node)
		}
		if event.Kind == "threshold_coin" || event.Kind == "prevote_no" || event.Kind == "finvote_no" || event.Kind == "prevote_yes" || event.Kind == "finvote_commit" || event.Kind == "halt" {
			fmt.Printf(" leader=%d", event.Leader)
		}
		if event.Value != "" {
			fmt.Printf(" value=%s", event.Value)
		}
		if event.Detail != "" {
			fmt.Printf(" %s", event.Detail)
		}
		fmt.Println()
	}

	fmt.Println("\nHonest-node decisions:")
	ids := make([]int, 0, len(result.Decisions))
	for id := range result.Decisions {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, id := range ids {
		fmt.Printf("  node %d -> %s\n", id, result.Decisions[smvba.NodeID(id)])
	}
}
