package resolver

import (
	"fmt"

	"github.com/okto-digital/regis3/internal/registry"
)

// Resolver handles dependency resolution for registry items.
type Resolver struct {
	graph    *Graph
	manifest *registry.Manifest
}

// NewResolver creates a resolver from a manifest.
func NewResolver(manifest *registry.Manifest) *Resolver {
	r := &Resolver{
		graph:    NewGraph(),
		manifest: manifest,
	}
	r.buildGraph()
	return r
}

// NewResolverFromItems creates a resolver from a list of items.
func NewResolverFromItems(items []*registry.Item) *Resolver {
	manifest := registry.NewManifest("")
	for _, item := range items {
		manifest.AddItem(item)
	}
	return NewResolver(manifest)
}

// buildGraph constructs the dependency graph from the manifest.
func (r *Resolver) buildGraph() {
	for _, item := range r.manifest.Items {
		r.graph.AddNode(
			item.FullName(),
			item.Type,
			item.Name,
			item.Deps,
		)
	}
}

// Graph returns the underlying dependency graph.
func (r *Resolver) Graph() *Graph {
	return r.graph
}

// ResolveResult contains the result of dependency resolution.
type ResolveResult struct {
	// Order is the installation order (dependencies first).
	Order []string

	// Items are the resolved items in installation order.
	Items []*registry.Item

	// Missing are dependencies that don't exist in the registry.
	Missing []string
}

// Resolve resolves dependencies for the given item IDs.
// Returns items in installation order (dependencies first).
func (r *Resolver) Resolve(ids []string) (*ResolveResult, error) {
	// Check for missing items
	for _, id := range ids {
		if _, ok := r.manifest.GetItem(id); !ok {
			return nil, fmt.Errorf("item not found: %s", id)
		}
	}

	// Get installation order
	order, err := r.graph.ResolveOrder(ids)
	if err != nil {
		return nil, err
	}

	// Collect items in order
	items := make([]*registry.Item, 0, len(order))
	for _, id := range order {
		if item, ok := r.manifest.GetItem(id); ok {
			items = append(items, item)
		}
	}

	// Find missing dependencies
	missing := r.findMissing(ids)

	return &ResolveResult{
		Order:   order,
		Items:   items,
		Missing: missing,
	}, nil
}

// ResolveAll resolves all items in the manifest.
// Returns items in installation order.
func (r *Resolver) ResolveAll() (*ResolveResult, error) {
	order, err := r.graph.TopologicalSort()
	if err != nil {
		return nil, err
	}

	items := make([]*registry.Item, 0, len(order))
	for _, id := range order {
		if item, ok := r.manifest.GetItem(id); ok {
			items = append(items, item)
		}
	}

	missing := r.graph.Validate()

	return &ResolveResult{
		Order:   order,
		Items:   items,
		Missing: missing,
	}, nil
}

// findMissing finds all missing dependencies for the given items.
func (r *Resolver) findMissing(ids []string) []string {
	seen := make(map[string]bool)
	var missing []string

	var check func(id string)
	check = func(id string) {
		item, ok := r.manifest.GetItem(id)
		if !ok {
			return
		}

		for _, dep := range item.Deps {
			if _, exists := r.manifest.GetItem(dep); !exists {
				if !seen[dep] {
					seen[dep] = true
					missing = append(missing, dep)
				}
			} else {
				check(dep)
			}
		}
	}

	for _, id := range ids {
		check(id)
	}

	return missing
}

// HasCycle checks if there are any circular dependencies.
func (r *Resolver) HasCycle() bool {
	return r.graph.HasCycle()
}

// FindCycle returns the cycle if one exists.
func (r *Resolver) FindCycle() ([]string, error) {
	_, err := r.graph.TopologicalSort()
	if err != nil {
		if cycleErr, ok := err.(*CycleError); ok {
			return cycleErr.Cycle, nil
		}
		return nil, err
	}
	return nil, nil
}

// DependencyInfo contains information about an item's dependencies.
type DependencyInfo struct {
	// ID is the item's full name.
	ID string

	// DirectDeps are the immediate dependencies.
	DirectDeps []string

	// AllDeps are all transitive dependencies.
	AllDeps []string

	// Dependents are items that depend on this item.
	Dependents []string

	// Missing are dependencies that don't exist.
	Missing []string
}

// GetDependencyInfo returns detailed dependency information for an item.
func (r *Resolver) GetDependencyInfo(id string) (*DependencyInfo, error) {
	item, ok := r.manifest.GetItem(id)
	if !ok {
		return nil, fmt.Errorf("item not found: %s", id)
	}

	// Find missing dependencies
	var missing []string
	for _, dep := range item.Deps {
		if _, exists := r.manifest.GetItem(dep); !exists {
			missing = append(missing, dep)
		}
	}

	return &DependencyInfo{
		ID:         id,
		DirectDeps: item.Deps,
		AllDeps:    r.graph.AllDependencies(id),
		Dependents: r.graph.Dependents(id),
		Missing:    missing,
	}, nil
}

// ValidateResult contains the result of dependency validation.
type ValidateResult struct {
	// Valid is true if all dependencies are satisfied and no cycles exist.
	Valid bool

	// Cycle contains the cycle path if one exists.
	Cycle []string

	// MissingDeps are dependencies that reference non-existent items.
	MissingDeps []string

	// Errors contains human-readable error messages.
	Errors []string
}

// Validate checks if all dependencies are valid.
func (r *Resolver) Validate() *ValidateResult {
	result := &ValidateResult{Valid: true}

	// Check for cycles
	if cycle, err := r.FindCycle(); err == nil && len(cycle) > 0 {
		result.Valid = false
		result.Cycle = cycle
		result.Errors = append(result.Errors,
			fmt.Sprintf("circular dependency: %v", cycle))
	}

	// Check for missing dependencies
	missing := r.graph.Validate()
	if len(missing) > 0 {
		result.Valid = false
		result.MissingDeps = missing
		for _, m := range missing {
			result.Errors = append(result.Errors, m)
		}
	}

	return result
}

// GetInstallOrder returns the installation order for specific items.
// This is a convenience method that returns just the order without full resolution.
func (r *Resolver) GetInstallOrder(ids []string) ([]string, error) {
	result, err := r.Resolve(ids)
	if err != nil {
		return nil, err
	}
	return result.Order, nil
}

// GetAllInstallOrder returns the installation order for all items.
func (r *Resolver) GetAllInstallOrder() ([]string, error) {
	return r.graph.TopologicalSort()
}

// FilterByType returns items of a specific type in dependency order.
func (r *Resolver) FilterByType(itemType string) ([]*registry.Item, error) {
	order, err := r.graph.TopologicalSort()
	if err != nil {
		return nil, err
	}

	var items []*registry.Item
	for _, id := range order {
		if item, ok := r.manifest.GetItem(id); ok {
			if item.Type == itemType {
				items = append(items, item)
			}
		}
	}

	return items, nil
}
