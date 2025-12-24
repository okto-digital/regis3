https://claude.ai/share/b589b7a4-0c08-443b-b535-83eab4879eeb


GO lang



Create a unified company repo with various skills, slash commands, claude.md precofingured files, subagents, etc. that could be initialised using a bash command okto-cc [parameters]

# regis3 - Complete Implementation Specification

## Table of Contents

1. [Overview](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#overview)
2. [Core Philosophy](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#core-philosophy)
3. [Registry Structure](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#registry-structure)
4. [YAML Frontmatter Schema](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#yaml-frontmatter-schema)
5. [Scanner & Manifest Builder](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#scanner--manifest-builder)
6. [CLI Commands](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#cli-commands)
7. [Output Modes](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#output-modes)
8. [Target System](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#target-system)
9. [Bootstrap & Self-Installation](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#bootstrap--self-installation)
10. [Go Implementation](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#go-implementation)
11. [AI Assist Mode (Future)](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#ai-assist-mode-future)
12. [Installation & Distribution](https://claude.ai/chat/31cb53a4-b63e-48df-844d-026ca18b11d6#installation--distribution)

---

## Overview

**regis3** is a configuration registry and CLI tool for managing LLM assistant configurations (skills, agents, commands, MCP configs). It supports multiple output targets (Claude Code, Cursor, GPT, etc.) from a single source of truth.

**Name origin:** regis (registry) + 3 (tri in Slovak) = "registry"

**Primary goals:**

- Centralized knowledge base for team conventions and AI instructions
- Self-describing files with YAML frontmatter
- Auto-discovery and manifest generation
- Multi-target output transformation
- Human-friendly and agent-friendly interfaces
- Self-installing capability (AI can install what it needs)

---

## Core Philosophy

|Principle|Description|
|---|---|
|**Single source of truth**|Each file contains its own metadata in YAML frontmatter|
|**Flat registry**|No enforced folder structure - organize however you want|
|**Computed manifest**|Scanner builds index from files, no manual maintenance|
|**Target-agnostic content**|Content is LLM-neutral, output adapts to target tool|
|**Agent-friendly**|JSON output, consistent schemas, exit codes for automation|
|**Self-aware**|Bootstrap skill enables AI to search and install what it needs|

---

## Registry Structure

### No Enforced Structure

Users organize files however they want. The only requirement is YAML frontmatter in `.md` files.

**Example user structures:**

```
# By domain
registry/
├── marketing/
│   ├── brand-voice.md
│   ├── content-writer.md
│   └── assets/
│       └── tone-guide.pdf
├── development/
│   ├── backend/
│   │   ├── node.md
│   │   └── python.md
│   ├── frontend/
│   │   └── vue.md
│   └── architect.md
└── devops/
    ├── deploy.md
    └── scripts/
        └── deploy.sh
```

```
# Flat
registry/
├── node-backend.md
├── vue-frontend.md
├── architect-agent.md
├── deploy-command.md
└── vue-bundle.md
```

```
# By type (traditional)
registry/
├── skills/
│   ├── node.md
│   └── vue.md
├── agents/
│   └── architect.md
└── bundles/
    └── vue.md
```

### Special Directories

```
registry/
├── meta/                       # Bootstrap and internal skills
│   └── regis3-bootstrap.md
├── targets/                    # Target definitions (optional)
│   ├── claude.yaml
│   ├── cursor.yaml
│   └── gpt.yaml
└── ... user content ...
```

### Generated Files

```
.build/                         # Gitignored
├── manifest.json               # Built from all YAML headers
├── validation.log              # Last build report
└── checksums.json              # For change detection (optional)
```

---

## YAML Frontmatter Schema

### Required Fields (All Types)

```yaml
---
type: skill | agent | command | mcp | template | bundle
name: unique-kebab-case-identifier
description: Clear description of what this does (10-20 words ideal)
---
```

### Optional Fields (All Types)

```yaml
---
deps:                           # Dependencies (auto-installed)
  - skill:git-conventions
  - skill:testing
tags:                           # Searchable keywords
  - javascript
  - backend
attachments:                    # Files that come with this item
  - assets/diagram.png
  - docs/api-spec.pdf
scripts:                        # Executable scripts
  - scripts/setup.sh
triggers:                       # Keywords for AI suggest command
  - node
  - express
  - REST API
target:                         # Per-target overrides
  cursor:
    exclude: true
  gpt:
    priority: low
---
```

### Type-Specific Fields

**Bundle:**

```yaml
---
type: bundle
name: vue-fullstack
description: Vue.js + Node.js full stack setup
template: vue-node              # Template to use for CLAUDE.md
items:                          # Required - items to install
  - skill:git-conventions
  - skill:node-backend
  - skill:vue-frontend
  - agent:architect
---
```

**MCP:**

````yaml
---
type: mcp
name: directus
description: Directus CMS integration
---

```json
{
  "mcpServers": {
    "directus": {
      "command": "npx",
      "args": ["-y", "@anthropic/mcp-directus"],
      "env": {
        "DIRECTUS_URL": "${DIRECTUS_URL}"
      }
    }
  }
}
````

````

### Complete Examples

**Skill:**
```yaml
---
type: skill
name: node-backend
description: Node.js + Express backend patterns and conventions
deps:
  - skill:git-conventions
  - skill:testing
tags:
  - javascript
  - node
  - express
  - backend
  - api
attachments:
  - assets/api-patterns.pdf
triggers:
  - node
  - express
  - REST
  - API
  - backend
---

# Node.js Backend Conventions

## Project Structure

...content...
````

**Agent:**

```yaml
---
type: agent
name: architect
description: System design, planning, and task breakdown
deps:
  - skill:git-conventions
tags:
  - planning
  - architecture
  - design
triggers:
  - design
  - architecture
  - plan
  - breakdown
---

# Architect Agent

You are responsible for system design...

...content...
```

**Command:**

```yaml
---
type: command
name: deploy
description: Deployment workflow and checklist
tags:
  - devops
  - deployment
scripts:
  - scripts/deploy.sh
  - scripts/rollback.sh
---

# /deploy Command

## Pre-deployment Checklist

...content...
```

---

## Scanner & Manifest Builder

### Build Command

```bash
./build.sh

# Or via regis3 CLI
regis3 build
```

### Scanner Responsibilities

1. **Discover files** - Find all `.md` files in registry
2. **Parse frontmatter** - Extract YAML from each file
3. **Validate required fields** - type, name, description
4. **Validate references** - deps, attachments, scripts exist
5. **Check uniqueness** - No duplicate type:name combinations
6. **Detect circular deps** - A → B → A
7. **Validate bundles** - All items in bundle exist
8. **Generate manifest** - Write `.build/manifest.json`
9. **Report issues** - Errors, warnings, suggestions

### Validation Rules

|Rule|Severity|Description|
|---|---|---|
|Missing `type`|Error|Required field|
|Missing `name`|Error|Required field|
|Missing `description`|Error|Required field|
|Invalid `type` value|Error|Must be: skill, agent, command, mcp, template, bundle|
|Duplicate `type:name`|Error|Names must be unique within type|
|Dep not found|Error|Referenced dependency doesn't exist|
|Circular dependency|Error|A → B → A detected|
|Bundle item not found|Error|Item in bundle doesn't exist|
|Attachment not found|Error|Listed file doesn't exist|
|Script not found|Error|Listed script doesn't exist|
|Missing `items` in bundle|Error|Bundles require items list|
|Short description|Warning|Less than 5 words|
|No tags|Warning|Tags recommended for searchability|
|Orphaned file|Warning|File not referenced anywhere|
|Similar content|Info|Possible duplicates detected|

### Build Output

**Success:**

```
$ regis3 build

Scanning registry...

  ✓ marketing/brand-voice.md
  ✓ marketing/content-writer.md
  ✓ development/backend/node.md
  ✓ development/frontend/vue.md
  ✓ development/architect.md
  ✓ bundles/vue-fullstack.md

Summary:
  ✓ 8 skills
  ✓ 3 agents
  ✓ 2 commands
  ✓ 1 mcp
  ✓ 2 templates
  ✓ 3 bundles

Build complete. Manifest written to .build/manifest.json
```

**With errors:**

```
$ regis3 build

Scanning registry...

  ✓ marketing/brand-voice.md
  ✗ marketing/new-file.md
    → missing required field: type
  ✗ development/backend/node.md
    → dependency not found: skill:nonexistent
  ⚠ development/frontend/vue.md
    → description very short (3 words)
  ⚠ assets/old-diagram.png
    → orphaned file (not referenced)

Found 2 errors, 2 warnings.
Exit code: 1
```

### Generated Manifest

```json
{
  "version": "1.0.0",
  "generated": "2025-01-15T10:30:00Z",
  "registry_path": "/Users/stefan/.regis3/registry",
  "items": {
    "skill:node-backend": {
      "type": "skill",
      "name": "node-backend",
      "source": "development/backend/node.md",
      "description": "Node.js + Express backend patterns",
      "deps": ["skill:git-conventions", "skill:testing"],
      "tags": ["javascript", "node", "backend"],
      "attachments": ["development/backend/assets/api-patterns.pdf"],
      "triggers": ["node", "express", "REST", "API"]
    },
    "agent:architect": {
      "type": "agent",
      "name": "architect",
      "source": "development/architect.md",
      "description": "System design and planning",
      "deps": ["skill:git-conventions"],
      "tags": ["planning", "architecture"],
      "triggers": ["design", "plan", "architecture"]
    }
  },
  "bundles": {
    "vue-fullstack": {
      "source": "bundles/vue-fullstack.md",
      "description": "Vue.js + Node.js full stack",
      "template": "vue-node",
      "items": ["skill:git-conventions", "skill:node-backend", "skill:vue-frontend", "agent:architect"]
    }
  },
  "templates": {
    "vue-node": {
      "source": "templates/vue-node.md",
      "description": "Vue.js + Node.js full stack template"
    }
  },
  "stats": {
    "skills": 8,
    "agents": 3,
    "commands": 2,
    "mcp": 1,
    "templates": 2,
    "bundles": 3
  }
}
```

---

## CLI Commands

### Quick Reference

```bash
# Bundles (shortcuts)
regis3 vue                          # Install Vue.js stack
regis3 rn                           # Install React Native stack
regis3 api                          # Install Node API stack

# Listing
regis3 list                         # All items by type
regis3 list skills                  # Just skills
regis3 list --tag backend           # Filter by tag

# Searching
regis3 search <query>               # Search name, description, tags
regis3 suggest "<task description>" # AI-powered recommendations

# Information
regis3 info skill:node-backend      # Detailed item info

# Installing
regis3 init                         # Bootstrap only (self-install mode)
regis3 init --template vue-node     # Bootstrap + template
regis3 add skill:node-backend       # Add item + dependencies
regis3 add skill:node skill:vue     # Add multiple items

# Status
regis3 status                       # Show installed items

# Building
regis3 build                        # Scan and build manifest
regis3 validate                     # Validate without building

# Maintenance
regis3 update                       # Git pull registry
regis3 orphans                      # List unreferenced files

# Help
regis3 help                         # General help
regis3 help <command>               # Command-specific help
regis3 version                      # Version info
```

### Command Details

#### `regis3 list`

```bash
regis3 list [type] [--tag TAG] [--json] [--quiet]

# Examples
regis3 list                     # All items grouped by type
regis3 list skills              # Just skills
regis3 list --tag marketing     # Items with marketing tag
regis3 list --json              # JSON output
regis3 list --quiet             # Just names (for piping)
```

#### `regis3 search`

```bash
regis3 search <query> [--type TYPE] [--limit N] [--json]

# Examples
regis3 search vue               # Search all fields
regis3 search vue --type skill  # Only skills
regis3 search "backend api"     # Multiple terms
regis3 search vue --json        # JSON output
```

#### `regis3 suggest`

```bash
regis3 suggest "<task description>" [--limit N] [--json]

# Examples
regis3 suggest "build a vue dashboard with charts"
regis3 suggest "REST API with authentication"

# Output
Based on your task, recommended items:

  skill:vue-frontend     Vue.js 3 patterns (match: vue, dashboard)
  skill:tailwind         Styling (match: dashboard)
  skill:node-backend     API patterns (match: API)
  skill:auth-jwt         JWT authentication (match: authentication)

Install all with:
  regis3 add skill:vue-frontend skill:tailwind skill:node-backend skill:auth-jwt
```

#### `regis3 info`

```bash
regis3 info <type:name> [--json]

# Example
regis3 info skill:node-backend

# Output
┌─────────────────────────────────────────┐
│ skill: node-backend                     │
└─────────────────────────────────────────┘

Description:  Node.js + Express backend patterns
Source:       development/backend/node.md

Dependencies:
  → skill:git-conventions
  → skill:testing

Tags:         javascript, node, backend, express, api

Attachments:
  → assets/api-patterns.pdf

Included in bundles:
  → vue-fullstack
  → api
```

#### `regis3 init`

```bash
regis3 init [--template NAME] [--target TARGET]

# Bootstrap only (enables self-installation)
regis3 init

# With template
regis3 init --template vue-node

# For different target
regis3 init --target cursor
```

#### `regis3 add`

```bash
regis3 add <type:name> [type:name...] [--target TARGET] [--force] [--json]

# Examples
regis3 add skill:node-backend           # Single item + deps
regis3 add skill:node skill:vue         # Multiple items
regis3 add agent:architect --force      # Overwrite if exists
regis3 add skill:node --target cursor   # Install for Cursor
```

#### `regis3 status`

```bash
regis3 status [--json]

# Output
Installed in this project:

  Skills:
    - git-conventions
    - node-backend
    - vue-frontend

  Agents:
    - architect

  Commands:
    (none)

  MCP:
    - directus

  Target: claude
  Installed: 2025-01-15
```

#### `regis3 build`

```bash
regis3 build [--strict] [--json]

# Options
--strict    Treat warnings as errors
--json      Output report as JSON
```

#### `regis3 validate`

```bash
regis3 validate [path] [--json]

# Validate entire registry
regis3 validate

# Validate single file
regis3 validate marketing/new-file.md
```

---

## Output Modes

### Human Mode (Default)

Pretty terminal output with colors, boxes, and formatting.

```bash
regis3 list

Skills:
  marketing/
    brand-voice          Marketing brand voice guidelines
    email-marketing      Email marketing best practices

  backend/
    node-backend         Node.js + Express patterns
    python               Python FastAPI patterns
```

### JSON Mode

Machine-readable output for agent integration.

```bash
regis3 list --json
```

```json
{
  "success": true,
  "command": "list",
  "data": {
    "items": [
      {
        "type": "skill",
        "name": "brand-voice",
        "description": "Marketing brand voice guidelines",
        "tags": ["marketing", "writing"]
      }
    ]
  },
  "messages": [],
  "error": null
}
```

### Quiet Mode

Just identifiers, for scripting and piping.

```bash
regis3 list --quiet

skill:brand-voice
skill:email-marketing
skill:node-backend
skill:python
agent:architect
```

### Consistent Response Schema

All commands return:

```json
{
  "success": true | false,
  "command": "command-name",
  "data": { ... } | null,
  "messages": [
    {"level": "info" | "warn" | "error", "text": "message"}
  ],
  "error": null | {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { ... }
  }
}
```

### Exit Codes

|Code|Meaning|
|---|---|
|0|Success|
|1|General error|
|2|Validation error|
|3|Not found|
|4|Dependency error|
|5|Already exists (optional)|

---

## Target System

### Concept

Content is LLM-agnostic. Target defines:

- Output file structure
- File naming conventions
- Content transformation

### Target Definition Files

**registry/targets/claude.yaml:**

```yaml
target: claude
description: Claude Code configuration

output:
  root_file: CLAUDE.md
  config_dir: .claude
  
structure:
  skill: "{{config_dir}}/skills/{{name}}/SKILL.md"
  agent: "{{config_dir}}/agents/{{name}}.md"
  command: "{{config_dir}}/commands/{{name}}.md"
  mcp: "{{config_dir}}/mcp/{{name}}.json"
  
transforms:
  mcp:
    extract_json: true    # Extract JSON from markdown body
```

**registry/targets/cursor.yaml:**

```yaml
target: cursor
description: Cursor IDE configuration

output:
  root_file: .cursorrules
  config_dir: null
  
structure:
  type: single_file       # Everything concatenated
  
template: |
  # Project Rules
  
  {{#each skills}}
  ## {{this.name}}
  {{this.content}}
  
  {{/each}}
  
  {{#each agents}}
  ## Agent: {{this.name}}
  {{this.content}}
  
  {{/each}}

transforms:
  mcp:
    exclude: true         # Cursor doesn't support MCP
```

**registry/targets/gpt.yaml:**

```yaml
target: gpt
description: Raw system prompt for GPT/ChatGPT

output:
  root_file: SYSTEM_PROMPT.md
  config_dir: null
  
structure:
  type: single_file
  max_tokens: 8000        # Trim if too long
  
template: |
  You are a senior developer. Follow these conventions:
  
  {{#each skills}}
  {{this.content}}
  ---
  {{/each}}

transforms:
  agent:
    exclude: true         # No multi-agent in raw GPT
  command:
    exclude: true
  mcp:
    exclude: true
```

### Output Examples

**Claude target:**

```
project/
├── CLAUDE.md
└── .claude/
    ├── skills/
    │   ├── git-conventions/
    │   │   └── SKILL.md
    │   └── node-backend/
    │       ├── SKILL.md
    │       └── assets/
    │           └── api-patterns.pdf
    ├── agents/
    │   └── architect.md
    └── mcp/
        └── directus.json
```

**Cursor target:**

```
project/
└── .cursorrules              # Single file with all content
```

**GPT target:**

```
project/
└── SYSTEM_PROMPT.md          # Single file, trimmed to token limit
```

### CLI Usage

```bash
# Default target (claude)
regis3 add skill:node-backend

# Explicit target
regis3 add skill:node-backend --target cursor

# Initialize for specific target
regis3 init --target cursor

# Export current project to different target
regis3 export cursor

# Set default target in config
regis3 config set default_target cursor
```

---

## Bootstrap & Self-Installation

### Concept

A meta-skill that teaches Claude (or other AI) how to use regis3. Once installed, the AI can search for and install skills as needed.

### Bootstrap Command

```bash
regis3 init

# Output
┌─────────────────────────────────────┐
│  regis3 initialized                 │
└─────────────────────────────────────┘

Created:
  ✓ CLAUDE.md (base template)
  ✓ .claude/skills/regis3-bootstrap/SKILL.md

Claude can now:
  • Search registry: regis3 search <query>
  • Get suggestions: regis3 suggest "<task>"
  • Install skills: regis3 add <type:name>
  • Check status: regis3 status

Or install a full stack:
  regis3 vue
  regis3 rn
  regis3 api
```

### Bootstrap Skill

**registry/meta/regis3-bootstrap.md:**

````yaml
---
type: skill
name: regis3-bootstrap
description: Enables self-installation of skills from regis3 registry
tags:
  - meta
  - bootstrap
  - core
priority: system
---

# regis3 Registry Access

You have access to a skill registry via the `regis3` CLI. Use it to find and install relevant skills for the current task.

## Available Commands

```bash
# Search for relevant skills
regis3 search <query>

# Get AI-powered recommendations
regis3 suggest "<task description>"

# List all available items
regis3 list
regis3 list --tag <tag>

# Get details about an item
regis3 info <type:name>

# Install items (with dependencies)
regis3 add <type:name> [<type:name> ...]

# See what's currently installed
regis3 status
````

## When to Use

Before starting a task, consider:

1. What domain is this task in? (frontend, backend, marketing, etc.)
2. What technologies are involved? (Vue, Node, Python, etc.)
3. Are there relevant skills in the registry?

Search the registry and install relevant skills BEFORE writing code.

## Example Workflow

User asks: "Help me write a Node.js API endpoint"

Your process:

1. Check current skills: `regis3 status`
2. If no Node.js skill: `regis3 search node` or `regis3 suggest "Node.js API"`
3. Review results and install: `regis3 add skill:node-backend`
4. Now proceed with the task using installed conventions

## Rules

- Always check if relevant skills exist before improvising
- Install skills that match the task domain
- Don't install everything - be selective
- If unsure, ask the user which skills to install
- Use `regis3 suggest` for complex tasks

```

### Self-Installation Flow

```

User: "I need to build a Vue.js dashboard with user authentication"

Claude: [Checks regis3 status - only bootstrap installed]

Claude: "Let me find the right skills for this task..."

[Runs: regis3 suggest "Vue.js dashboard with user authentication"]

Claude: "I found these relevant skills:

- skill:vue-frontend (Vue.js patterns)
- skill:node-backend (API patterns)
- skill:auth-jwt (Authentication)

Should I install them?"

User: "Yes"

[Runs: regis3 add skill:vue-frontend skill:node-backend skill:auth-jwt]

Claude: "Done. Now I have the conventions for: ✓ Vue.js component patterns ✓ Node.js API structure ✓ JWT authentication

Let's start building your dashboard..."

```

---

## Go Implementation

### Project Structure

```

regis3/ ├── cmd/ │ └── regis3/ │ └── main.go # Entry point │ ├── internal/ │ ├── cli/ # CLI commands │ │ ├── root.go # Root command & global flags │ │ ├── list.go │ │ ├── search.go │ │ ├── suggest.go │ │ ├── info.go │ │ ├── init.go │ │ ├── add.go │ │ ├── status.go │ │ ├── build.go │ │ ├── validate.go │ │ ├── update.go │ │ ├── export.go │ │ ├── config.go │ │ └── assist.go # AI assist commands (future) │ │ │ ├── registry/ # Registry operations │ │ ├── scanner.go # Scan files, parse frontmatter │ │ ├── manifest.go # Build/read manifest │ │ ├── item.go # Item struct & methods │ │ ├── bundle.go # Bundle struct & methods │ │ └── validator.go # Validation rules │ │ │ ├── resolver/ # Dependency resolution │ │ ├── resolver.go # Main resolver │ │ └── graph.go # Dependency graph │ │ │ ├── installer/ # Installation │ │ ├── installer.go # Copy files to project │ │ ├── targets.go # Load target definitions │ │ └── transform.go # Content transformation │ │ │ ├── output/ # Output formatting │ │ ├── writer.go # Writer interface │ │ ├── json.go # JSON writer │ │ ├── pretty.go # Human-friendly writer │ │ └── quiet.go # Minimal writer │ │ │ ├── config/ # Configuration │ │ ├── config.go # App configuration │ │ └── paths.go # Path resolution │ │ │ └── assist/ # AI assist (future) │ ├── analyzer.go │ ├── claude.go │ └── suggestions.go │ ├── pkg/ │ └── frontmatter/ # Reusable frontmatter parser │ ├── parser.go │ └── parser_test.go │ ├── targets/ # Built-in target definitions │ ├── claude.yaml │ ├── cursor.yaml │ └── gpt.yaml │ ├── registry/ # Default/example registry │ ├── meta/ │ │ └── regis3-bootstrap.md │ └── examples/ │ └── ... │ ├── go.mod ├── go.sum ├── Makefile ├── .goreleaser.yaml └── README.md

````

### Dependencies

```go
// go.mod
module github.com/okto-digital/regis3

go 1.21

require (
    // CLI framework
    github.com/spf13/cobra v1.8.0
    github.com/spf13/viper v1.18.0
    
    // YAML parsing
    gopkg.in/yaml.v3 v3.0.1
    
    // Terminal UI (Charm stack)
    github.com/charmbracelet/lipgloss v0.9.1
    github.com/charmbracelet/bubbles v0.18.0
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/huh v0.3.0
    
    // File operations
    github.com/spf13/afero v1.11.0
    github.com/otiai10/copy v1.14.0
    github.com/bmatcuk/doublestar/v4 v4.6.1
    
    // Validation
    github.com/go-playground/validator/v10 v10.17.0
    
    // Errors
    go.uber.org/multierr v1.11.0
    
    // Testing
    github.com/stretchr/testify v1.8.4
)
````

### Key Interfaces

```go
// internal/output/writer.go
package output

type Writer interface {
    Success(data interface{})
    Error(err error)
    Info(msg string)
    Warn(msg string)
    Progress(msg string)
    Table(headers []string, rows [][]string)
}

func NewWriter(format string) Writer {
    switch format {
    case "json":
        return &JSONWriter{}
    case "quiet":
        return &QuietWriter{}
    default:
        return &PrettyWriter{}
    }
}
```

```go
// internal/registry/item.go
package registry

type Item struct {
    Type        string   `yaml:"type" json:"type" validate:"required,oneof=skill agent command mcp template bundle"`
    Name        string   `yaml:"name" json:"name" validate:"required"`
    Description string   `yaml:"description" json:"description" validate:"required"`
    Deps        []string `yaml:"deps" json:"deps,omitempty"`
    Tags        []string `yaml:"tags" json:"tags,omitempty"`
    Attachments []string `yaml:"attachments" json:"attachments,omitempty"`
    Scripts     []string `yaml:"scripts" json:"scripts,omitempty"`
    Triggers    []string `yaml:"triggers" json:"triggers,omitempty"`
    Target      map[string]TargetOverride `yaml:"target" json:"target,omitempty"`
    
    // Computed fields
    Source      string   `json:"source"`
    Content     string   `json:"-"`
}

func (i *Item) FullName() string {
    return fmt.Sprintf("%s:%s", i.Type, i.Name)
}

type TargetOverride struct {
    Exclude  bool   `yaml:"exclude" json:"exclude,omitempty"`
    Priority string `yaml:"priority" json:"priority,omitempty"`
}
```

```go
// internal/registry/manifest.go
package registry

type Manifest struct {
    Version      string              `json:"version"`
    Generated    time.Time           `json:"generated"`
    RegistryPath string              `json:"registry_path"`
    Items        map[string]*Item    `json:"items"`
    Bundles      map[string]*Bundle  `json:"bundles"`
    Templates    map[string]*Template `json:"templates"`
    Stats        Stats               `json:"stats"`
}

type Bundle struct {
    Source      string   `json:"source"`
    Description string   `json:"description"`
    Template    string   `json:"template,omitempty"`
    Items       []string `json:"items"`
}

type Stats struct {
    Skills    int `json:"skills"`
    Agents    int `json:"agents"`
    Commands  int `json:"commands"`
    MCP       int `json:"mcp"`
    Templates int `json:"templates"`
    Bundles   int `json:"bundles"`
}
```

```go
// internal/resolver/resolver.go
package resolver

type Resolver interface {
    // Resolve returns items in installation order (deps first)
    Resolve(items []string) ([]*registry.Item, error)
    
    // Check for circular dependencies
    ValidateDeps() error
}

type resolver struct {
    manifest *registry.Manifest
}

func New(m *registry.Manifest) Resolver {
    return &resolver{manifest: m}
}
```

```go
// internal/installer/installer.go
package installer

type Installer interface {
    Install(items []*registry.Item, opts InstallOptions) error
    Status() (*ProjectStatus, error)
}

type InstallOptions struct {
    Target    string
    Dest      string
    Force     bool
    DryRun    bool
}

type ProjectStatus struct {
    Target    string   `json:"target"`
    Installed []string `json:"installed"`
    Date      string   `json:"date"`
}
```

### Example Command Implementation

```go
// internal/cli/add.go
package cli

import (
    "github.com/spf13/cobra"
    "github.com/okto-digital/regis3/internal/output"
    "github.com/okto-digital/regis3/internal/registry"
    "github.com/okto-digital/regis3/internal/resolver"
    "github.com/okto-digital/regis3/internal/installer"
)

func newAddCmd() *cobra.Command {
    var target string
    var force bool
    
    cmd := &cobra.Command{
        Use:   "add <type:name> [type:name...]",
        Short: "Add items to current project",
        Args:  cobra.MinimumNArgs(1),
        RunE: func(cmd *cobra.Command, args []string) error {
            w := output.NewWriter(outputFormat)
            
            // Load manifest
            m, err := registry.LoadManifest()
            if err != nil {
                w.Error(err)
                return err
            }
            
            // Resolve dependencies
            r := resolver.New(m)
            items, err := r.Resolve(args)
            if err != nil {
                w.Error(err)
                return err
            }
            
            // Install
            inst := installer.New(m)
            err = inst.Install(items, installer.InstallOptions{
                Target: target,
                Force:  force,
            })
            if err != nil {
                w.Error(err)
                return err
            }
            
            // Output
            w.Success(map[string]interface{}{
                "installed": itemNames(items),
            })
            
            return nil
        },
    }
    
    cmd.Flags().StringVarP(&target, "target", "t", "claude", "Target platform")
    cmd.Flags().BoolVarP(&force, "force", "f", false, "Overwrite existing files")
    
    return cmd
}
```

---

## AI Assist Mode (Future)

### Commands

```bash
# Analyze single file - suggest YAML header
regis3 assist analyze <file>

# Scan entire registry - find issues
regis3 assist scan

# Auto-fix with AI suggestions
regis3 assist fix <file>

# Interactive repair session
regis3 assist repair

# Best practices review
regis3 assist review <file>
```

### Use Cases

|Scenario|Command|What AI Does|
|---|---|---|
|New file without header|`assist analyze`|Reads content, suggests complete YAML|
|Malformed YAML|`assist fix`|Repairs syntax, fills missing fields|
|Poor description|`assist review`|Suggests better description|
|Missing deps|`assist scan`|Detects likely dependencies from content|
|Inconsistent tags|`assist scan`|Suggests tag normalization|
|Duplicate content|`assist scan`|Finds similar files, suggests merge|

### Example Flows

**Analyze new file:**

```bash
$ regis3 assist analyze marketing/new-file.md

Analyzing: marketing/new-file.md

Content summary:
  - Appears to be a skill (instructional content)
  - Topic: Email marketing best practices
  - References: brand-voice (likely dependency)
  - ~450 words

Suggested YAML header:

---
type: skill
name: email-marketing
description: Email marketing best practices and templates
deps:
  - skill:brand-voice
tags:
  - marketing
  - email
  - copywriting
---

Apply this header? (y/n/edit)
```

**Scan registry:**

```bash
$ regis3 assist scan

Scanning registry...

Issues found:

  Missing YAML header:
    ⚠ marketing/new-file.md
    ⚠ development/notes.md

  Incomplete header:
    ⚠ backend/node.md - missing description
    ⚠ agents/writer.md - missing tags (optional but recommended)

  Potential issues:
    ? frontend/vue.md - description very short (5 words)
    ? devops/deploy.md - might depend on skill:git (referenced in content)

  Possible duplicates:
    ? backend/node.md and backend/nodejs.md (78% similar)

  Orphaned files:
    ? assets/old-diagram.png - not referenced anywhere

Run 'regis3 assist fix <file>' to repair individual files
Run 'regis3 assist repair' for interactive repair session
```

**Interactive repair:**

```bash
$ regis3 assist repair

Starting interactive repair...

[1/3] marketing/new-file.md (missing header)

Content preview:
  # Email Marketing Best Practices
  This guide covers how to write effective...

AI suggests:
  type: skill
  name: email-marketing
  description: Email marketing best practices and templates
  deps: [skill:brand-voice]
  tags: [marketing, email]

Actions:
  a) Apply suggested header
  e) Edit before applying
  s) Skip this file
  q) Quit

> a

✓ Applied header to marketing/new-file.md

[2/3] backend/node.md (missing description)
...
```

### Maintainer Agent

**registry/meta/regis3-maintainer.md:**

````yaml
---
type: agent
name: regis3-maintainer
description: AI agent for registry maintenance and content improvement
tags:
  - meta
  - internal
---

# regis3 Registry Maintainer

You help maintain the regis3 registry by:

1. Analyzing markdown files and suggesting YAML frontmatter
2. Reviewing existing content for quality and completeness
3. Detecting missing dependencies by analyzing content
4. Suggesting tag normalization across the registry
5. Finding duplicate or similar content

## YAML Schema

When suggesting headers, always use this format:

```yaml
---
type: skill | agent | command | mcp | template | bundle
name: unique-kebab-case-name
description: Clear, concise description (10-20 words)
deps:
  - type:name
tags:
  - relevant
  - searchable
  - terms
---
````

## Guidelines

- name: kebab-case, unique, descriptive
- description: What it does, not how. 10-20 words ideal.
- deps: Only direct dependencies, not transitive
- tags: 3-7 tags, lowercase, common terms

## Detecting Dependencies

Look for:

- Explicit references: "see skill:xyz" or "uses xyz"
- Implicit references: "following our git workflow" → skill:git
- Code imports: mentions of frameworks/libraries

## Quality Checks

- Description should be actionable
- Content should have clear structure
- Code examples should be present for technical skills
- No orphaned references

````

### CLI Flags

```bash
regis3 assist analyze <file>
  --apply           # Auto-apply without confirmation
  --dry-run         # Show what would be done
  --model claude    # Which AI model (default: claude)
  --json            # Output as JSON

regis3 assist scan
  --fix             # Auto-fix obvious issues
  --strict          # Treat warnings as errors
  --json            # Output as JSON

regis3 assist repair
  --auto            # Accept all AI suggestions
  --skip-confirm    # Don't ask for confirmation
````

---

## Installation & Distribution

### Build Commands

```makefile
# Makefile
VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-s -w -X main.version=$(VERSION)"

.PHONY: build
build:
	go build $(LDFLAGS) -o bin/regis3 ./cmd/regis3

.PHONY: install
install:
	go install $(LDFLAGS) ./cmd/regis3

.PHONY: release
release:
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o dist/regis3-darwin-amd64 ./cmd/regis3
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o dist/regis3-darwin-arm64 ./cmd/regis3
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o dist/regis3-linux-amd64 ./cmd/regis3
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o dist/regis3-linux-arm64 ./cmd/regis3
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o dist/regis3-windows-amd64.exe ./cmd/regis3

.PHONY: test
test:
	go test -v ./...

.PHONY: lint
lint:
	golangci-lint run
```

### GoReleaser Config

```yaml
# .goreleaser.yaml
project_name: regis3

builds:
  - main: ./cmd/regis3
    binary: regis3
    env:
      - CGO_ENABLED=0
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}}

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

changelog:
  sort: asc

brews:
  - name: regis3
    repository:
      owner: okto-digital
      name: homebrew-tap
    homepage: https://github.com/okto-digital/regis3
    description: Configuration registry for LLM assistants
    install: |
      bin.install "regis3"
```

### User Installation

**Option 1: Download binary**

```bash
# macOS (Apple Silicon)
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_darwin_arm64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_darwin_amd64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/

# Linux
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_linux_amd64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/

# Windows (PowerShell)
Invoke-WebRequest -Uri "https://github.com/okto-digital/regis3/releases/latest/download/regis3_windows_amd64.zip" -OutFile regis3.zip
Expand-Archive regis3.zip
Move-Item regis3/regis3.exe C:\Users\$env:USERNAME\bin\
```

**Option 2: Homebrew (macOS/Linux)**

```bash
brew tap okto-digital/tap
brew install regis3
```

**Option 3: Go install**

```bash
go install github.com/okto-digital/regis3/cmd/regis3@latest
```

### Registry Setup

```bash
# Clone your registry
git clone git@github.com:okto-digital/regis3-registry.git ~/.regis3/registry

# Or set custom path
regis3 config set registry_path /path/to/registry

# Build manifest
regis3 build
```

### Configuration File

```yaml
# ~/.regis3/config.yaml
registry_path: ~/.regis3/registry
default_target: claude
output_format: pretty   # pretty, json, quiet
```

---

## Summary

|Component|Description|
|---|---|
|**Registry**|Flat file structure, user-organized, YAML frontmatter|
|**Scanner**|Builds manifest from files, validates, reports issues|
|**CLI**|Human & agent friendly, multiple output modes|
|**Targets**|Claude, Cursor, GPT, extensible|
|**Bootstrap**|Self-installation capability for AI|
|**AI Assist**|Future: automated maintenance and repair|
|**Language**|Go with Cobra + Charm stack|
|**Distribution**|Cross-platform binaries, Homebrew, go install|

---

_Document version: 1.0_ _Last updated: January 2025_



Good. Let me compile the final specification with everything we've discussed:

---

# regis3 - Final Implementation Specification v2

## Overview

**regis3** is a configuration registry and CLI tool for managing LLM assistant configurations. It features a flat, user-organized registry with YAML frontmatter, automatic manifest generation, and multi-target output support.

**Name origin:** regis (registry) + 3 (tri in Slovak) = "registry"

---

## Core Principles

|Principle|Description|
|---|---|
|**Single source of truth**|YAML frontmatter in each file|
|**Flat registry**|No enforced structure - organize your way|
|**Namespaced YAML**|All metadata under `regis3:` key to avoid conflicts|
|**Computed manifest**|Scanner builds index automatically|
|**Multi-target**|Same content, different output formats|
|**Self-aware**|Bootstrap enables AI to self-install|

---

## Registry Structure

No enforced structure. Users organize files however they want:

```
registry/
├── marketing/
│   ├── brand-voice.md
│   ├── content-writer.md
│   └── assets/
│       └── tone-guide.pdf
│
├── development/
│   ├── backend/
│   │   ├── node.md
│   │   └── python.md
│   ├── frontend/
│   │   └── vue.md
│   └── architect.md
│
├── philosophies/
│   ├── clean-code.md
│   └── simple-easy.md
│
├── stacks/
│   ├── vue.md
│   └── api.md
│
├── meta/
│   └── regis3-bootstrap.md
│
├── targets/
│   ├── claude.yaml
│   ├── cursor.yaml
│   └── gpt.yaml
│
└── .build/                     # Generated (gitignored)
    └── manifest.json
```

---

## YAML Schema

All metadata under `regis3:` namespace to avoid conflicts with Claude or other tools.

### Compact Format (Recommended)

```yaml
---
regis3:
  type: skill
  name: node-backend
  desc: Node.js + Express patterns
  cat: backend
  deps: [skill:git, skill:testing]
  tags: [node, javascript, express, api]
---
```

### Full Schema

```yaml
---
regis3:
  # Required
  type: skill | subagent | command | mcp | script | doc | project | philosophy | ruleset | stack | hook | prompt
  name: kebab-case-id
  desc: Short description (10-20 words)

  # Optional (all types)
  cat: category/subcategory       # For listing hierarchy
  deps: [type:name]               # Dependencies
  tags: [search, terms]           # Search keywords
  files: [relative/path]          # Related files (relative to this file)
  status: draft | stable | deprecated
  author: name
  order: 10                       # For merge types (lower = earlier in CLAUDE.md)
  target:                         # Per-target overrides
    cursor: {exclude: true}
    gpt: {exclude: true}

  # Hook-only
  trigger: pre-install | post-install | pre-build | post-build
  run: path/to/script.sh
---

Content here...
```

---

## Types (12 Total)

### Install Types (Separate Files)

|Type|Output Location|Description|
|---|---|---|
|`skill`|`.claude/skills/{name}/SKILL.md`|Conventions, patterns|
|`subagent`|`.claude/agents/{name}.md`|Worker agent definitions|
|`command`|`.claude/commands/{name}.md`|Slash commands|
|`mcp`|`.claude/mcp/{name}.json`|MCP server configs|
|`script`|`.claude/scripts/{name}.sh`|Utility scripts|
|`doc`|`.claude/docs/{name}.md`|Reference documentation|

### Merge Types (→ CLAUDE.md)

|Type|Purpose|Typical Order|
|---|---|---|
|`project`|Base project template|1-10|
|`philosophy`|High-level principles|10-30|
|`ruleset`|Specific rules|40-60|

### Meta Types

|Type|Purpose|
|---|---|
|`stack`|Composes deps into project (installs deps + merges CLAUDE.md)|
|`hook`|Lifecycle scripts (pre/post install/build)|
|`prompt`|Export only (for GPT, etc.)|

---

## Type Examples

### Skill

```yaml
---
regis3:
  type: skill
  name: node-backend
  desc: Node.js + Express patterns
  cat: backend
  deps: [skill:git, skill:testing]
  tags: [node, javascript, express, api]
  files: [assets/api-patterns.pdf]
---

# Node.js Backend Conventions

## Project Structure
...
```

### Subagent

```yaml
---
regis3:
  type: subagent
  name: architect
  desc: System design and task planning
  deps: [skill:git]
  tags: [planning, design]
---

# Architect Agent

You are responsible for system design...
```

### Philosophy

```yaml
---
regis3:
  type: philosophy
  name: clean-code
  desc: Clean code principles
  order: 10
  tags: [principles, quality]
---

## Philosophy: Clean Code

- Write code for humans first, computers second
- Single responsibility principle
- Meaningful names over comments
```

### Ruleset

```yaml
---
regis3:
  type: ruleset
  name: typescript-strict
  desc: Strict TypeScript rules
  order: 50
  tags: [typescript, rules]
---

## Rules: TypeScript

- Always use strict mode
- No `any` types
- Explicit return types on public functions
```

### Project

```yaml
---
regis3:
  type: project
  name: vue-base
  desc: Vue.js project base template
  order: 1
---

# Project: {{name}}

## Overview
Vue.js + Node.js full stack application.

## Tech Stack
- Frontend: Vue.js 3, Pinia, Vue Router
- Backend: Node.js, Express, Prisma
```

### Stack

```yaml
---
regis3:
  type: stack
  name: vue
  desc: Vue.js full stack
  deps:
    # Merge into CLAUDE.md (in order)
    - project:vue-base
    - philosophy:clean-code
    - philosophy:simple-easy
    - ruleset:typescript-strict
    # Install to .claude/
    - skill:git
    - skill:node-backend
    - skill:vue
    - subagent:architect
---
```

### MCP

````yaml
---
regis3: {type: mcp, name: directus, desc: Directus CMS integration}
---

```json
{"mcpServers":{"directus":{"command":"npx","args":["-y","@anthropic/mcp-directus"]}}}
````

````

### Hook
```yaml
---
regis3:
  type: hook
  name: setup-env
  desc: Creates .env from template
  trigger: post-install
  run: scripts/setup-env.sh
---
````

---

## File References

Paths are **relative to the .md file's directory**:

```
registry/
├── development/
│   ├── backend/
│   │   ├── node.md                 # ← This file
│   │   ├── assets/
│   │   │   └── patterns.pdf        # files: [assets/patterns.pdf]
│   │   └── scripts/
│   │       └── setup.sh            # files: [scripts/setup.sh]
```

Use `../` for shared files or use deps for shared resources.

---

## Scanner & Manifest Builder

### Build Command

```bash
regis3 build
```

### Validation Rules

|Rule|Severity|
|---|---|
|Missing `type`, `name`, `desc`|Error|
|Invalid `type` value|Error|
|Duplicate `type:name`|Error|
|Dep not found|Error|
|Circular dependency|Error|
|File in `files` not found|Error|
|Missing `regis3:` block|Error|
|Short description (<5 words)|Warning|
|No tags|Warning|
|Orphaned files|Warning|

### Generated Manifest

```json
{
  "version": "1.0.0",
  "generated": "2025-01-15T10:30:00Z",
  "items": {
    "skill:node-backend": {
      "type": "skill",
      "name": "node-backend",
      "source": "development/backend/node.md",
      "desc": "Node.js + Express patterns",
      "cat": "backend",
      "deps": ["skill:git", "skill:testing"],
      "tags": ["node", "javascript"],
      "files": ["development/backend/assets/patterns.pdf"]
    }
  },
  "stacks": {
    "vue": {
      "source": "stacks/vue.md",
      "desc": "Vue.js full stack",
      "deps": ["project:vue-base", "skill:git", "skill:vue"]
    }
  }
}
```

---

## CLI Commands

```bash
# Stacks (shortcuts)
regis3 vue                          # Install Vue stack
regis3 api                          # Install API stack

# Listing
regis3 list                         # All items by type
regis3 list skills                  # Just skills
regis3 list --cat backend           # Filter by category
regis3 list --tag javascript        # Filter by tag

# Searching
regis3 search <query>               # Search name, desc, tags
regis3 suggest "<task>"             # AI-powered recommendations

# Information
regis3 info skill:node-backend      # Detailed item info

# Installing
regis3 init                         # Bootstrap only
regis3 init --stack vue             # Bootstrap + stack
regis3 add skill:node-backend       # Add item + deps
regis3 add skill:node skill:vue     # Add multiple

# Status
regis3 status                       # Show installed items

# Building
regis3 build                        # Scan and build manifest
regis3 validate                     # Validate only
regis3 orphans                      # List unreferenced files

# Maintenance
regis3 update                       # Git pull registry

# AI Assist (future)
regis3 assist analyze <file>        # Suggest YAML header
regis3 assist scan                  # Find issues
regis3 assist repair                # Interactive fix
```

---

## Output Modes

```bash
regis3 list                  # Human-friendly (default)
regis3 list --json           # Machine-readable
regis3 list --quiet          # Just names
```

### JSON Response Schema

```json
{
  "success": true,
  "command": "list",
  "data": { ... },
  "messages": [{"level": "info", "text": "..."}],
  "error": null
}
```

### Exit Codes

|Code|Meaning|
|---|---|
|0|Success|
|1|General error|
|2|Validation error|
|3|Not found|
|4|Dependency error|

---

## Target System

Targets define output format. Content is LLM-agnostic.

**targets/claude.yaml:**

```yaml
target: claude
output:
  root_file: CLAUDE.md
  config_dir: .claude
structure:
  skill: "{{config_dir}}/skills/{{name}}/SKILL.md"
  subagent: "{{config_dir}}/agents/{{name}}.md"
  command: "{{config_dir}}/commands/{{name}}.md"
  mcp: "{{config_dir}}/mcp/{{name}}.json"
```

**targets/cursor.yaml:**

```yaml
target: cursor
output:
  root_file: .cursorrules
structure:
  type: single_file
```

---

## Bootstrap & Self-Installation

```bash
regis3 init
```

Installs bootstrap skill that teaches Claude to:

```bash
regis3 search <query>
regis3 suggest "<task>"
regis3 add <type:name>
regis3 status
```

---

## Go Implementation

### Dependencies

- CLI: `github.com/spf13/cobra`, `github.com/spf13/viper`
- YAML: `gopkg.in/yaml.v3`
- UI: `github.com/charmbracelet/lipgloss`, `bubbletea`, `huh`
- Files: `github.com/spf13/afero`, `github.com/otiai10/copy`

### Project Structure

```
regis3/
├── cmd/regis3/main.go
├── internal/
│   ├── cli/           # Commands
│   ├── registry/      # Scanner, manifest, validation
│   ├── resolver/      # Dependency resolution
│   ├── installer/     # File copying, transforms
│   ├── output/        # JSON, pretty, quiet writers
│   └── assist/        # AI assist (future)
├── pkg/frontmatter/   # YAML parser
├── targets/           # Target definitions
└── registry/          # Example content
```

---

## AI Assist Mode (Future)

```bash
regis3 assist analyze <file>    # Suggest YAML header
regis3 assist scan              # Find all issues
regis3 assist fix <file>        # Auto-repair
regis3 assist repair            # Interactive session
regis3 assist review <file>     # Best practices check
```

---

## Distribution

- **Binary downloads:** GitHub releases (darwin, linux, windows)
- **Homebrew:** `brew install okto-digital/tap/regis3`
- **Go install:** `go install github.com/okto-digital/regis3@latest`

---

_Specification version: 2.0_ _Last updated: January 2025_

---

Ready to build! 🚀