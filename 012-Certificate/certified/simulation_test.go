package certified

import "testing"

func TestSimulationCertifiesHonestBlocks(t *testing.T) {
	result, err := RunSimulation(4, 1, 2, 2)
	if err != nil {
		t.Fatal(err)
	}
	// Node 0 is simulated as faulty/silent; nodes 1,2,3 should each certify two blocks.
	if got, want := len(result.Certificates), 6; got != want {
		t.Fatalf("certificates = %d, want %d", got, want)
	}
	for _, frontier := range result.Frontiers {
		if frontier[0].Height != 0 {
			t.Fatalf("faulty proposer unexpectedly certified height %d", frontier[0].Height)
		}
		for proposer := NodeID(1); proposer <= 3; proposer++ {
			if frontier[proposer].Height != 2 {
				t.Fatalf("proposer %d height = %d, want 2", proposer, frontier[proposer].Height)
			}
		}
	}
}

func TestCommitteeRejectsUnsafeFaultBound(t *testing.T) {
	if _, err := RunSimulation(4, 2, 1, 1); err == nil {
		t.Fatal("expected n >= 3f+1 validation error")
	}
}
