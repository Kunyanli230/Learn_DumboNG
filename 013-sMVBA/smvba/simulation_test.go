package smvba

import "testing"

func TestSimulatorDecidesSameValueForAllHonestNodes(t *testing.T) {
	sim, err := NewSimulator(4, 1, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	result, err := sim.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.Decided.Proposer != 1 {
		t.Fatalf("decided proposer = %d, want 1 because leader 0 is faulty and leader 1 is first honest", result.Decided.Proposer)
	}
	want := result.Decided.Digest()
	for id, got := range result.Decisions {
		if got != want {
			t.Fatalf("node %d decided %s, want %s", id, got, want)
		}
	}
	if got, wantCount := len(result.Decisions), 3; got != wantCount {
		t.Fatalf("honest decisions = %d, want %d", got, wantCount)
	}
}

func TestSimulatorRejectsUnsafeFaultBound(t *testing.T) {
	if _, err := NewSimulator(4, 2, 0, 0); err == nil {
		t.Fatal("expected n >= 3f+1 validation error")
	}
}
