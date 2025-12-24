package resolver

import (
	"testing"

	"github.com/okto-digital/regis3/internal/registry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestItems() []*registry.Item {
	return []*registry.Item{
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "philosophy",
				Name: "clean-code",
				Desc: "Clean code principles",
			},
			Source: "philosophies/clean-code.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "git-conventions",
				Desc: "Git conventions",
			},
			Source: "skills/git-conventions.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "testing",
				Desc: "Testing practices",
				Deps: []string{"skill:git-conventions"},
			},
			Source: "skills/testing.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "subagent",
				Name: "architect",
				Desc: "Architecture agent",
				Deps: []string{"skill:git-conventions"},
			},
			Source: "agents/architect.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "stack",
				Name: "base",
				Desc: "Base stack",
				Deps: []string{
					"philosophy:clean-code",
					"skill:git-conventions",
					"skill:testing",
					"subagent:architect",
				},
			},
			Source: "stacks/base.md",
		},
	}
}

func TestNewResolver(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	assert.NotNil(t, r.Graph())
	assert.Len(t, r.Graph().Nodes(), 5)
}

func TestResolver_Resolve(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	result, err := r.Resolve([]string{"stack:base"})
	require.NoError(t, err)

	assert.Len(t, result.Order, 5)
	assert.Len(t, result.Items, 5)

	// stack:base should be last
	assert.Equal(t, "stack:base", result.Order[len(result.Order)-1])

	// git-conventions should come before testing
	indexGit := indexOf(result.Order, "skill:git-conventions")
	indexTesting := indexOf(result.Order, "skill:testing")
	assert.Less(t, indexGit, indexTesting)
}

func TestResolver_Resolve_SingleItem(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	result, err := r.Resolve([]string{"skill:testing"})
	require.NoError(t, err)

	// Should include testing and its dependency (git-conventions)
	assert.Len(t, result.Order, 2)
	assert.Contains(t, result.Order, "skill:testing")
	assert.Contains(t, result.Order, "skill:git-conventions")
}

func TestResolver_Resolve_NotFound(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	_, err := r.Resolve([]string{"skill:nonexistent"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestResolver_ResolveAll(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	result, err := r.ResolveAll()
	require.NoError(t, err)

	assert.Len(t, result.Order, 5)
	assert.Len(t, result.Items, 5)
	assert.Empty(t, result.Missing)
}

func TestResolver_HasCycle(t *testing.T) {
	// Create items with a cycle
	items := []*registry.Item{
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "a",
				Desc: "Skill A",
				Deps: []string{"skill:b"},
			},
			Source: "a.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "b",
				Desc: "Skill B",
				Deps: []string{"skill:a"},
			},
			Source: "b.md",
		},
	}

	r := NewResolverFromItems(items)
	assert.True(t, r.HasCycle())
}

func TestResolver_FindCycle(t *testing.T) {
	items := []*registry.Item{
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "a",
				Desc: "Skill A",
				Deps: []string{"skill:b"},
			},
			Source: "a.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "b",
				Desc: "Skill B",
				Deps: []string{"skill:c"},
			},
			Source: "b.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "c",
				Desc: "Skill C",
				Deps: []string{"skill:a"},
			},
			Source: "c.md",
		},
	}

	r := NewResolverFromItems(items)
	cycle, err := r.FindCycle()
	require.NoError(t, err)
	assert.NotEmpty(t, cycle)
}

func TestResolver_FindCycle_NoCycle(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	cycle, err := r.FindCycle()
	require.NoError(t, err)
	assert.Empty(t, cycle)
}

func TestResolver_GetDependencyInfo(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	info, err := r.GetDependencyInfo("skill:testing")
	require.NoError(t, err)

	assert.Equal(t, "skill:testing", info.ID)
	assert.Equal(t, []string{"skill:git-conventions"}, info.DirectDeps)
	assert.Equal(t, []string{"skill:git-conventions"}, info.AllDeps)
	assert.Empty(t, info.Missing)
}

func TestResolver_GetDependencyInfo_Dependents(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	info, err := r.GetDependencyInfo("skill:git-conventions")
	require.NoError(t, err)

	// git-conventions is depended on by testing, architect, and base
	assert.Len(t, info.Dependents, 3)
	assert.Contains(t, info.Dependents, "skill:testing")
	assert.Contains(t, info.Dependents, "subagent:architect")
	assert.Contains(t, info.Dependents, "stack:base")
}

func TestResolver_GetDependencyInfo_NotFound(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	_, err := r.GetDependencyInfo("skill:nonexistent")
	require.Error(t, err)
}

func TestResolver_GetDependencyInfo_MissingDeps(t *testing.T) {
	items := []*registry.Item{
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "test",
				Desc: "Test skill",
				Deps: []string{"skill:missing"},
			},
			Source: "test.md",
		},
	}

	r := NewResolverFromItems(items)
	info, err := r.GetDependencyInfo("skill:test")
	require.NoError(t, err)

	assert.Equal(t, []string{"skill:missing"}, info.Missing)
}

func TestResolver_Validate(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	result := r.Validate()
	assert.True(t, result.Valid)
	assert.Empty(t, result.Cycle)
	assert.Empty(t, result.MissingDeps)
	assert.Empty(t, result.Errors)
}

func TestResolver_Validate_WithCycle(t *testing.T) {
	items := []*registry.Item{
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "a",
				Desc: "Skill A",
				Deps: []string{"skill:b"},
			},
			Source: "a.md",
		},
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "b",
				Desc: "Skill B",
				Deps: []string{"skill:a"},
			},
			Source: "b.md",
		},
	}

	r := NewResolverFromItems(items)
	result := r.Validate()

	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.Cycle)
	assert.NotEmpty(t, result.Errors)
}

func TestResolver_Validate_WithMissingDeps(t *testing.T) {
	items := []*registry.Item{
		{
			Regis3Meta: registry.Regis3Meta{
				Type: "skill",
				Name: "test",
				Desc: "Test skill",
				Deps: []string{"skill:missing"},
			},
			Source: "test.md",
		},
	}

	r := NewResolverFromItems(items)
	result := r.Validate()

	assert.False(t, result.Valid)
	assert.NotEmpty(t, result.MissingDeps)
	assert.NotEmpty(t, result.Errors)
}

func TestResolver_GetInstallOrder(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	order, err := r.GetInstallOrder([]string{"skill:testing"})
	require.NoError(t, err)

	assert.Len(t, order, 2)
	assert.Equal(t, "skill:git-conventions", order[0])
	assert.Equal(t, "skill:testing", order[1])
}

func TestResolver_GetAllInstallOrder(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	order, err := r.GetAllInstallOrder()
	require.NoError(t, err)

	assert.Len(t, order, 5)
}

func TestResolver_FilterByType(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	skills, err := r.FilterByType("skill")
	require.NoError(t, err)

	assert.Len(t, skills, 2)
	for _, skill := range skills {
		assert.Equal(t, "skill", skill.Type)
	}
}

func TestResolver_FilterByType_Ordered(t *testing.T) {
	items := createTestItems()
	r := NewResolverFromItems(items)

	skills, err := r.FilterByType("skill")
	require.NoError(t, err)

	// git-conventions should come before testing due to dependency
	var gitIndex, testingIndex int
	for i, skill := range skills {
		if skill.Name == "git-conventions" {
			gitIndex = i
		}
		if skill.Name == "testing" {
			testingIndex = i
		}
	}

	assert.Less(t, gitIndex, testingIndex)
}

func TestResolver_WithManifest(t *testing.T) {
	manifest := registry.NewManifest("/test")

	for _, item := range createTestItems() {
		manifest.AddItem(item)
	}

	r := NewResolver(manifest)

	result, err := r.ResolveAll()
	require.NoError(t, err)

	assert.Len(t, result.Items, 5)
}

func TestResolver_ComplexChain(t *testing.T) {
	// A -> B -> C -> D -> E
	items := []*registry.Item{
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "e", Desc: "E"}, Source: "e.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "d", Desc: "D", Deps: []string{"skill:e"}}, Source: "d.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "c", Desc: "C", Deps: []string{"skill:d"}}, Source: "c.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "b", Desc: "B", Deps: []string{"skill:c"}}, Source: "b.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "a", Desc: "A", Deps: []string{"skill:b"}}, Source: "a.md"},
	}

	r := NewResolverFromItems(items)

	order, err := r.GetInstallOrder([]string{"skill:a"})
	require.NoError(t, err)

	assert.Len(t, order, 5)
	assert.Equal(t, []string{"skill:e", "skill:d", "skill:c", "skill:b", "skill:a"}, order)
}

func TestResolver_DiamondDependency(t *testing.T) {
	//     A
	//    / \
	//   B   C
	//    \ /
	//     D
	items := []*registry.Item{
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "a", Desc: "A"}, Source: "a.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "b", Desc: "B", Deps: []string{"skill:a"}}, Source: "b.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "c", Desc: "C", Deps: []string{"skill:a"}}, Source: "c.md"},
		{Regis3Meta: registry.Regis3Meta{Type: "skill", Name: "d", Desc: "D", Deps: []string{"skill:b", "skill:c"}}, Source: "d.md"},
	}

	r := NewResolverFromItems(items)

	order, err := r.GetInstallOrder([]string{"skill:d"})
	require.NoError(t, err)

	// A should be first, D should be last
	assert.Equal(t, "skill:a", order[0])
	assert.Equal(t, "skill:d", order[len(order)-1])

	// B and C should be between A and D
	indexB := indexOf(order, "skill:b")
	indexC := indexOf(order, "skill:c")
	assert.Greater(t, indexB, 0)
	assert.Greater(t, indexC, 0)
	assert.Less(t, indexB, len(order)-1)
	assert.Less(t, indexC, len(order)-1)
}
