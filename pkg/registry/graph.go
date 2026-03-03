package registry

import (
	"fmt"
	"sort"

	"ghost-ops/pkg/protocol"
)

// ServiceGraph represents a dependency graph of services.
// Edges are directed from dependent to dependency (Service -> Dependency).
// It supports cycle detection and topological sorting.
type ServiceGraph struct {
	nodes map[string][]string
}

// NewServiceGraph creates a graph from existing service records.
func NewServiceGraph(services []protocol.ServiceRecord) *ServiceGraph {
	g := &ServiceGraph{
		nodes: make(map[string][]string),
	}
	for _, svc := range services {
		g.nodes[svc.ServiceID] = svc.Dependencies
	}
	return g
}

// AddService adds a service and its dependencies to the graph.
// It returns an error if adding the service creates a cycle.
func (g *ServiceGraph) AddService(serviceID string, dependencies []string) error {
	// Temporarily add the node
	originalDeps, exists := g.nodes[serviceID]
	g.nodes[serviceID] = dependencies

	// Check for cycles
	if g.HasCycle() {
		// Revert
		if exists {
			g.nodes[serviceID] = originalDeps
		} else {
			delete(g.nodes, serviceID)
		}
		return fmt.Errorf("dependency cycle detected for service %s", serviceID)
	}
	return nil
}

// HasCycle checks if the graph contains any cycles.
func (g *ServiceGraph) HasCycle() bool {
	visited := make(map[string]bool)
	recursionStack := make(map[string]bool)

	// Sort keys for deterministic behavior
	keys := make([]string, 0, len(g.nodes))
	for k := range g.nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, node := range keys {
		if !visited[node] {
			if g.hasCycleDFS(node, visited, recursionStack) {
				return true
			}
		}
	}
	return false
}

func (g *ServiceGraph) hasCycleDFS(node string, visited, recursionStack map[string]bool) bool {
	visited[node] = true
	recursionStack[node] = true

	for _, dep := range g.nodes[node] {
		// Self-loop check
		if dep == node {
			return true
		}

		if _, known := g.nodes[dep]; !known {
			continue
		}

		if !visited[dep] {
			if g.hasCycleDFS(dep, visited, recursionStack) {
				return true
			}
		} else if recursionStack[dep] {
			return true
		}
	}

	recursionStack[node] = false
	return false
}

// TopologicalSort returns a valid build order (dependencies first).
// Since edges are Dependent -> Dependency, a DFS post-order traversal yields:
// [Dependencies] ... [Dependents]
func (g *ServiceGraph) TopologicalSort() ([]string, error) {
	if g.HasCycle() {
		return nil, fmt.Errorf("cannot sort: graph has cycles")
	}

	visited := make(map[string]bool)
	order := make([]string, 0)

	// Sort keys for deterministic output
	keys := make([]string, 0, len(g.nodes))
	for k := range g.nodes {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, node := range keys {
		if !visited[node] {
			g.visit(node, visited, &order)
		}
	}

	return order, nil
}

func (g *ServiceGraph) visit(node string, visited map[string]bool, order *[]string) {
	visited[node] = true
	for _, dep := range g.nodes[node] {
		// Only traverse known nodes in the graph
		if _, known := g.nodes[dep]; known && !visited[dep] {
			g.visit(dep, visited, order)
		}
	}
	*order = append(*order, node)
}
