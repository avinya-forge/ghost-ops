package registry

import (
	"testing"

	"ghost-ops/pkg/protocol"
)

func TestServiceGraph_CycleDetection(t *testing.T) {
	tests := []struct {
		name       string
		existing   []protocol.ServiceRecord
		newService string
		newDeps    []string
		wantError  bool
	}{
		{
			name:       "No Cycle - Empty Graph",
			existing:   nil,
			newService: "A",
			newDeps:    []string{"B"},
			wantError:  false,
		},
		{
			name: "No Cycle - Linear",
			existing: []protocol.ServiceRecord{
				{ServiceID: "B", Dependencies: []string{"C"}},
			},
			newService: "A",
			newDeps:    []string{"B"},
			wantError:  false,
		},
		{
			name: "Cycle - Direct",
			existing: []protocol.ServiceRecord{
				{ServiceID: "B", Dependencies: []string{"A"}},
			},
			newService: "A",
			newDeps:    []string{"B"},
			wantError:  true,
		},
		{
			name: "Cycle - Indirect",
			existing: []protocol.ServiceRecord{
				{ServiceID: "B", Dependencies: []string{"C"}},
				{ServiceID: "C", Dependencies: []string{"A"}},
			},
			newService: "A",
			newDeps:    []string{"B"},
			wantError:  true,
		},
		{
			name:       "Cycle - Self Loop",
			existing:   nil,
			newService: "A",
			newDeps:    []string{"A"},
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewServiceGraph(tt.existing)
			err := g.AddService(tt.newService, tt.newDeps)
			if (err != nil) != tt.wantError {
				t.Errorf("AddService() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestServiceGraph_TopologicalSort(t *testing.T) {
	services := []protocol.ServiceRecord{
		{ServiceID: "A", Dependencies: []string{"B"}},
		{ServiceID: "B", Dependencies: []string{"C"}},
		{ServiceID: "C", Dependencies: []string{}},
		{ServiceID: "D", Dependencies: []string{"A"}},
	}
	// Expected Order: C, B, A, D
	// C has no deps. B depends on C. A depends on B. D depends on A.

	g := NewServiceGraph(services)
	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort() error = %v", err)
	}

	// Verify order constraints
	pos := make(map[string]int)
	for i, s := range order {
		pos[s] = i
	}

	// Check dependencies come before dependents
	if pos["B"] < pos["C"] {
		t.Errorf("B should come after C")
	}
	if pos["A"] < pos["B"] {
		t.Errorf("A should come after B")
	}
	if pos["D"] < pos["A"] {
		t.Errorf("D should come after A")
	}

	if len(order) != 4 {
		t.Errorf("Expected 4 items, got %d", len(order))
	}
}

func TestServiceGraph_HasCycle_PreExisting(t *testing.T) {
	services := []protocol.ServiceRecord{
		{ServiceID: "A", Dependencies: []string{"B"}},
		{ServiceID: "B", Dependencies: []string{"A"}},
	}
	g := NewServiceGraph(services)
	if !g.HasCycle() {
		t.Error("Expected cycle detection on pre-existing cyclic graph")
	}
}
