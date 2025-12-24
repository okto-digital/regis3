package resolver

import (
	"fmt"
	"sort"
	"strings"
)

// Graph represents a directed dependency graph.
type Graph struct {
	nodes map[string]*Node
	edges map[string][]string // node -> dependencies
}

// Node represents an item in the dependency graph.
type Node struct {
	ID   string // Full name (e.g., "skill:git-conventions")
	Type string
	Name string
	Deps []string
}

// NewGraph creates an empty dependency graph.
func NewGraph() *Graph {
	return &Graph{
		nodes: make(map[string]*Node),
		edges: make(map[string][]string),
	}
}

// AddNode adds a node to the graph.
func (g *Graph) AddNode(id, itemType, name string, deps []string) {
	g.nodes[id] = &Node{
		ID:   id,
		Type: itemType,
		Name: name,
		Deps: deps,
	}
	g.edges[id] = deps
}

// GetNode returns a node by ID.
func (g *Graph) GetNode(id string) (*Node, bool) {
	node, ok := g.nodes[id]
	return node, ok
}

// Nodes returns all node IDs.
func (g *Graph) Nodes() []string {
	ids := make([]string, 0, len(g.nodes))
	for id := range g.nodes {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	return ids
}

// Dependencies returns the direct dependencies of a node.
func (g *Graph) Dependencies(id string) []string {
	return g.edges[id]
}

// Dependents returns all nodes that depend on the given node.
func (g *Graph) Dependents(id string) []string {
	var dependents []string
	for nodeID, deps := range g.edges {
		for _, dep := range deps {
			if dep == id {
				dependents = append(dependents, nodeID)
				break
			}
		}
	}
	sort.Strings(dependents)
	return dependents
}

// CycleError represents a circular dependency error.
type CycleError struct {
	Cycle []string
}

func (e *CycleError) Error() string {
	return fmt.Sprintf("circular dependency detected: %s", strings.Join(e.Cycle, " -> "))
}

// TopologicalSort returns nodes in dependency order (dependencies first).
// Returns an error if a cycle is detected.
func (g *Graph) TopologicalSort() ([]string, error) {
	// Kahn's algorithm for topological sorting
	// Also naturally detects cycles

	// Calculate in-degree for each node (number of dependencies)
	inDegree := make(map[string]int)
	for id := range g.nodes {
		inDegree[id] = 0
	}

	// Build reverse edges (who depends on whom)
	reverseDeps := make(map[string][]string)
	for id, deps := range g.edges {
		for _, dep := range deps {
			// Only count dependencies that exist in the graph
			if _, exists := g.nodes[dep]; exists {
				inDegree[id]++
				reverseDeps[dep] = append(reverseDeps[dep], id)
			}
		}
	}

	// Find all nodes with no dependencies (in-degree 0)
	var queue []string
	for id, degree := range inDegree {
		if degree == 0 {
			queue = append(queue, id)
		}
	}
	sort.Strings(queue) // Ensure deterministic order

	var result []string
	for len(queue) > 0 {
		// Take first node from queue
		node := queue[0]
		queue = queue[1:]
		result = append(result, node)

		// For each node that depends on this one
		dependents := reverseDeps[node]
		sort.Strings(dependents)
		for _, dependent := range dependents {
			inDegree[dependent]--
			if inDegree[dependent] == 0 {
				queue = append(queue, dependent)
				sort.Strings(queue)
			}
		}
	}

	// If we didn't process all nodes, there's a cycle
	if len(result) != len(g.nodes) {
		cycle := g.findCycle()
		return nil, &CycleError{Cycle: cycle}
	}

	return result, nil
}

// findCycle finds a cycle in the graph using DFS.
func (g *Graph) findCycle() []string {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)
	parent := make(map[string]string)

	var cycle []string
	var dfs func(node string) bool

	dfs = func(node string) bool {
		visited[node] = true
		recStack[node] = true

		for _, dep := range g.edges[node] {
			// Only consider dependencies that exist in the graph
			if _, exists := g.nodes[dep]; !exists {
				continue
			}

			if !visited[dep] {
				parent[dep] = node
				if dfs(dep) {
					return true
				}
			} else if recStack[dep] {
				// Found a cycle - reconstruct it
				cycle = []string{dep}
				for curr := node; curr != dep; curr = parent[curr] {
					cycle = append([]string{curr}, cycle...)
				}
				cycle = append([]string{dep}, cycle...)
				return true
			}
		}

		recStack[node] = false
		return false
	}

	// Try DFS from each unvisited node
	nodes := g.Nodes()
	for _, node := range nodes {
		if !visited[node] {
			if dfs(node) {
				return cycle
			}
		}
	}

	return nil
}

// HasCycle checks if the graph contains any cycles.
func (g *Graph) HasCycle() bool {
	_, err := g.TopologicalSort()
	return err != nil
}

// AllDependencies returns all transitive dependencies of a node.
func (g *Graph) AllDependencies(id string) []string {
	visited := make(map[string]bool)
	var result []string

	var visit func(nodeID string)
	visit = func(nodeID string) {
		for _, dep := range g.edges[nodeID] {
			if _, exists := g.nodes[dep]; !exists {
				continue
			}
			if !visited[dep] {
				visited[dep] = true
				result = append(result, dep)
				visit(dep)
			}
		}
	}

	visit(id)
	sort.Strings(result)
	return result
}

// ResolveOrder returns the installation order for a set of items.
// It includes all transitive dependencies and returns them in order.
func (g *Graph) ResolveOrder(ids []string) ([]string, error) {
	// Collect all required nodes (requested + all dependencies)
	required := make(map[string]bool)
	for _, id := range ids {
		required[id] = true
		for _, dep := range g.AllDependencies(id) {
			required[dep] = true
		}
	}

	// Create a subgraph with only required nodes
	subgraph := NewGraph()
	for id := range required {
		if node, ok := g.nodes[id]; ok {
			// Only include deps that are in our required set
			var filteredDeps []string
			for _, dep := range node.Deps {
				if required[dep] {
					filteredDeps = append(filteredDeps, dep)
				}
			}
			subgraph.AddNode(node.ID, node.Type, node.Name, filteredDeps)
		}
	}

	return subgraph.TopologicalSort()
}

// Validate checks if all dependencies reference existing nodes.
func (g *Graph) Validate() []string {
	var missing []string
	seen := make(map[string]bool)

	for id, deps := range g.edges {
		for _, dep := range deps {
			if _, exists := g.nodes[dep]; !exists {
				key := fmt.Sprintf("%s -> %s", id, dep)
				if !seen[key] {
					seen[key] = true
					missing = append(missing, fmt.Sprintf("%s depends on missing: %s", id, dep))
				}
			}
		}
	}

	sort.Strings(missing)
	return missing
}
