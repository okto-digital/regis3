package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGraph_AddNode(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:base", "skill", "base", nil)
	g.AddNode("skill:advanced", "skill", "advanced", []string{"skill:base"})

	node, ok := g.GetNode("skill:base")
	require.True(t, ok)
	assert.Equal(t, "skill:base", node.ID)
	assert.Equal(t, "skill", node.Type)
	assert.Equal(t, "base", node.Name)

	node, ok = g.GetNode("skill:advanced")
	require.True(t, ok)
	assert.Equal(t, []string{"skill:base"}, node.Deps)
}

func TestGraph_Nodes(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:c", "skill", "c", nil)
	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", nil)

	nodes := g.Nodes()
	assert.Equal(t, []string{"skill:a", "skill:b", "skill:c"}, nodes)
}

func TestGraph_Dependencies(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:base", "skill", "base", nil)
	g.AddNode("skill:advanced", "skill", "advanced", []string{"skill:base"})

	deps := g.Dependencies("skill:base")
	assert.Empty(t, deps)

	deps = g.Dependencies("skill:advanced")
	assert.Equal(t, []string{"skill:base"}, deps)
}

func TestGraph_Dependents(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:base", "skill", "base", nil)
	g.AddNode("skill:mid", "skill", "mid", []string{"skill:base"})
	g.AddNode("skill:top", "skill", "top", []string{"skill:base", "skill:mid"})

	dependents := g.Dependents("skill:base")
	assert.Equal(t, []string{"skill:mid", "skill:top"}, dependents)

	dependents = g.Dependents("skill:mid")
	assert.Equal(t, []string{"skill:top"}, dependents)

	dependents = g.Dependents("skill:top")
	assert.Empty(t, dependents)
}

func TestGraph_TopologicalSort_Simple(t *testing.T) {
	g := NewGraph()

	// A -> B -> C (C depends on B, B depends on A)
	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", []string{"skill:b"})

	order, err := g.TopologicalSort()
	require.NoError(t, err)

	// A must come before B, B must come before C
	indexA := indexOf(order, "skill:a")
	indexB := indexOf(order, "skill:b")
	indexC := indexOf(order, "skill:c")

	assert.Less(t, indexA, indexB, "A should come before B")
	assert.Less(t, indexB, indexC, "B should come before C")
}

func TestGraph_TopologicalSort_Diamond(t *testing.T) {
	g := NewGraph()

	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", []string{"skill:a"})
	g.AddNode("skill:d", "skill", "d", []string{"skill:b", "skill:c"})

	order, err := g.TopologicalSort()
	require.NoError(t, err)

	indexA := indexOf(order, "skill:a")
	indexB := indexOf(order, "skill:b")
	indexC := indexOf(order, "skill:c")
	indexD := indexOf(order, "skill:d")

	assert.Less(t, indexA, indexB, "A should come before B")
	assert.Less(t, indexA, indexC, "A should come before C")
	assert.Less(t, indexB, indexD, "B should come before D")
	assert.Less(t, indexC, indexD, "C should come before D")
}

func TestGraph_TopologicalSort_NoDeps(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", nil)
	g.AddNode("skill:c", "skill", "c", nil)

	order, err := g.TopologicalSort()
	require.NoError(t, err)

	// With no dependencies, should be alphabetical order
	assert.Equal(t, []string{"skill:a", "skill:b", "skill:c"}, order)
}

func TestGraph_TopologicalSort_Cycle(t *testing.T) {
	g := NewGraph()

	// A -> B -> C -> A (cycle)
	g.AddNode("skill:a", "skill", "a", []string{"skill:c"})
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", []string{"skill:b"})

	order, err := g.TopologicalSort()
	assert.Nil(t, order)
	require.Error(t, err)

	cycleErr, ok := err.(*CycleError)
	require.True(t, ok, "should be a CycleError")
	assert.NotEmpty(t, cycleErr.Cycle)
	assert.Contains(t, cycleErr.Error(), "circular dependency")
}

func TestGraph_TopologicalSort_SelfCycle(t *testing.T) {
	g := NewGraph()

	// A depends on itself
	g.AddNode("skill:a", "skill", "a", []string{"skill:a"})

	order, err := g.TopologicalSort()
	assert.Nil(t, order)
	require.Error(t, err)

	cycleErr, ok := err.(*CycleError)
	require.True(t, ok, "should be a CycleError")
	assert.Contains(t, cycleErr.Cycle, "skill:a")
}

func TestGraph_HasCycle(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(*Graph)
		hasCycle bool
	}{
		{
			name: "no cycle",
			setup: func(g *Graph) {
				g.AddNode("a", "skill", "a", nil)
				g.AddNode("b", "skill", "b", []string{"a"})
			},
			hasCycle: false,
		},
		{
			name: "simple cycle",
			setup: func(g *Graph) {
				g.AddNode("a", "skill", "a", []string{"b"})
				g.AddNode("b", "skill", "b", []string{"a"})
			},
			hasCycle: true,
		},
		{
			name: "self cycle",
			setup: func(g *Graph) {
				g.AddNode("a", "skill", "a", []string{"a"})
			},
			hasCycle: true,
		},
		{
			name: "long cycle",
			setup: func(g *Graph) {
				g.AddNode("a", "skill", "a", []string{"b"})
				g.AddNode("b", "skill", "b", []string{"c"})
				g.AddNode("c", "skill", "c", []string{"d"})
				g.AddNode("d", "skill", "d", []string{"a"})
			},
			hasCycle: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewGraph()
			tt.setup(g)
			assert.Equal(t, tt.hasCycle, g.HasCycle())
		})
	}
}

func TestGraph_AllDependencies(t *testing.T) {
	g := NewGraph()

	// A -> B -> C -> D
	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", []string{"skill:b"})
	g.AddNode("skill:d", "skill", "d", []string{"skill:c"})

	// D should transitively depend on A, B, C
	allDeps := g.AllDependencies("skill:d")
	assert.ElementsMatch(t, []string{"skill:a", "skill:b", "skill:c"}, allDeps)

	// B should only depend on A
	allDeps = g.AllDependencies("skill:b")
	assert.Equal(t, []string{"skill:a"}, allDeps)

	// A has no dependencies
	allDeps = g.AllDependencies("skill:a")
	assert.Empty(t, allDeps)
}

func TestGraph_AllDependencies_Diamond(t *testing.T) {
	g := NewGraph()

	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", []string{"skill:a"})
	g.AddNode("skill:d", "skill", "d", []string{"skill:b", "skill:c"})

	allDeps := g.AllDependencies("skill:d")
	assert.ElementsMatch(t, []string{"skill:a", "skill:b", "skill:c"}, allDeps)
}

func TestGraph_ResolveOrder(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", []string{"skill:b"})
	g.AddNode("skill:d", "skill", "d", nil) // Unrelated

	// Request only C - should include A and B as dependencies
	order, err := g.ResolveOrder([]string{"skill:c"})
	require.NoError(t, err)

	assert.Len(t, order, 3)
	assert.NotContains(t, order, "skill:d") // D should not be included

	indexA := indexOf(order, "skill:a")
	indexB := indexOf(order, "skill:b")
	indexC := indexOf(order, "skill:c")

	assert.Less(t, indexA, indexB)
	assert.Less(t, indexB, indexC)
}

func TestGraph_ResolveOrder_Multiple(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})
	g.AddNode("skill:c", "skill", "c", nil)
	g.AddNode("skill:d", "skill", "d", []string{"skill:c"})

	// Request B and D - should include A, B, C, D
	order, err := g.ResolveOrder([]string{"skill:b", "skill:d"})
	require.NoError(t, err)

	assert.Len(t, order, 4)
}

func TestGraph_ResolveOrder_WithCycle(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:a", "skill", "a", []string{"skill:b"})
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})

	_, err := g.ResolveOrder([]string{"skill:a"})
	require.Error(t, err)

	_, ok := err.(*CycleError)
	assert.True(t, ok)
}

func TestGraph_Validate(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a", "skill:missing"})

	missing := g.Validate()
	require.Len(t, missing, 1)
	assert.Contains(t, missing[0], "skill:missing")
}

func TestGraph_Validate_AllPresent(t *testing.T) {
	g := NewGraph()

	g.AddNode("skill:a", "skill", "a", nil)
	g.AddNode("skill:b", "skill", "b", []string{"skill:a"})

	missing := g.Validate()
	assert.Empty(t, missing)
}

func TestGraph_ComplexScenario(t *testing.T) {
	g := NewGraph()

	// Simulate a real-world scenario:
	// - philosophy:clean-code (no deps)
	// - skill:git-conventions (no deps)
	// - skill:testing (depends on git-conventions)
	// - subagent:architect (depends on git-conventions)
	// - stack:base (depends on all above)

	g.AddNode("philosophy:clean-code", "philosophy", "clean-code", nil)
	g.AddNode("skill:git-conventions", "skill", "git-conventions", nil)
	g.AddNode("skill:testing", "skill", "testing", []string{"skill:git-conventions"})
	g.AddNode("subagent:architect", "subagent", "architect", []string{"skill:git-conventions"})
	g.AddNode("stack:base", "stack", "base", []string{
		"philosophy:clean-code",
		"skill:git-conventions",
		"skill:testing",
		"subagent:architect",
	})

	// Should have no cycles
	assert.False(t, g.HasCycle())

	// Get installation order for stack:base
	order, err := g.ResolveOrder([]string{"stack:base"})
	require.NoError(t, err)

	assert.Len(t, order, 5)

	// stack:base should be last
	assert.Equal(t, "stack:base", order[len(order)-1])

	// Dependencies should come before dependents
	indexGit := indexOf(order, "skill:git-conventions")
	indexTesting := indexOf(order, "skill:testing")
	indexArchitect := indexOf(order, "subagent:architect")

	assert.Less(t, indexGit, indexTesting, "git-conventions before testing")
	assert.Less(t, indexGit, indexArchitect, "git-conventions before architect")
}

func TestCycleError_Error(t *testing.T) {
	err := &CycleError{Cycle: []string{"a", "b", "c", "a"}}
	assert.Contains(t, err.Error(), "circular dependency")
	assert.Contains(t, err.Error(), "a -> b -> c -> a")
}

// Helper function
func indexOf(slice []string, item string) int {
	for i, s := range slice {
		if s == item {
			return i
		}
	}
	return -1
}
