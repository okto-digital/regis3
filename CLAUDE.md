# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**regis3** - A Go CLI tool for managing LLM assistant configurations (skills, agents, commands, MCP configs). It maintains a single source of truth for team conventions with YAML frontmatter metadata, auto-discovery, and multi-target output (Claude, Cursor, GPT).

**Status:** Specification phase. The `PROJECT-regis3.md` contains the complete implementation spec. No Go source code exists yet.

## Build Commands

```bash
# Build (when implemented)
make build              # Build to bin/regis3
make install            # Install to $GOPATH/bin
make release            # Cross-platform builds to dist/

# Test
go test -v ./...        # Run all tests
go test -v ./pkg/...    # Test specific package
go test -cover ./...    # With coverage
go test -race ./...     # Race detector

# Lint/Format
gofmt -w .              # Format code
golangci-lint run       # Run linters

# Development
go run ./cmd/regis3 <command>   # Run without building
```

## Planned Architecture

```
cmd/regis3/main.go              # Entry point
internal/
  cli/                          # Cobra commands (list, search, add, build, etc.)
  registry/                     # Scanner, manifest builder, validation
  resolver/                     # Dependency resolution, cycle detection
  installer/                    # File copying, target transforms
  output/                       # Writers: JSON, pretty, quiet
  config/                       # App configuration, path resolution
pkg/frontmatter/                # Reusable YAML frontmatter parser
targets/                        # Target definitions (claude.yaml, cursor.yaml, gpt.yaml)
```

## Key Dependencies (Planned)

- CLI: `github.com/spf13/cobra`, `github.com/spf13/viper`
- YAML: `gopkg.in/yaml.v3`
- Terminal UI: `github.com/charmbracelet/lipgloss`, `bubbletea`, `huh`
- Files: `github.com/spf13/afero`, `github.com/otiai10/copy`
- Testing: `github.com/stretchr/testify`

## Development Environment

- Go 1.25 (trixie) via devcontainer
- GoReleaser for distribution

## Core Design Patterns

- **Interface-based design:** Accept interfaces, return structs
- **Functional options:** For complex configuration
- **Table-driven tests:** With subtests for organization
- **Error wrapping:** Always add context to errors
- **Exit codes:** 0=success, 1=general, 2=validation, 3=not found, 4=dependency error

## YAML Frontmatter Schema

All registry items use `regis3:` namespace:

```yaml
---
regis3:
  type: skill | subagent | command | mcp | script | doc | project | philosophy | ruleset | stack | hook | prompt
  name: kebab-case-identifier
  desc: Clear description (10-20 words)
  deps: [type:name, ...]        # Optional
  tags: [search, keywords]      # Optional
  order: 10                     # For merge types (lower = earlier in CLAUDE.md)
---
```

## Summary Principles

1. **Make it work, make it right, make it fast** (in that order)
2. **Small is beautiful** - small functions, small files
3. **Express intent** - code should read like prose
4. **One thing** - functions do one thing, classes have one responsibility
5. **DRY** - every piece of knowledge has one authoritative representation
6. **YAGNI** - don't add functionality until needed
7. **Boy Scout Rule** - leave code cleaner than you found it
8. **Tests are documentation** - show exactly how code should be used
9. **Continuous improvement** - refactor relentlessly
10. **Professionalism** - clean code is professional survival

Full clean code philosophy: `.claude/docs/clean-code-principles.md`
