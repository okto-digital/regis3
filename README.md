# regis3

A registry manager for LLM assistant configurations. Manage and install skills, agents, commands, and other configurations for tools like Claude Code, Cursor, and more.

## What is regis3?

regis3 is a **package manager for Claude Code configurations**.

You maintain a **registry** (a central folder of skills, agents, philosophies, etc.) and use regis3 to **install items from that registry into your current project**.

```
┌─────────────────┐       regis3 project add        ┌─────────────────┐
│    Registry     │  ─────────────────────────────▶ │  Your Project   │
│ ~/.regis3/      │                                 │  .claude/       │
│   skills/       │                                 │    skills/      │
│   agents/       │                                 │    agents/      │
│   commands/     │                                 │    commands/    │
└─────────────────┘                                 └─────────────────┘
```

**Target**: Claude Code (installs to `.claude/` directory)

## Features

- **Registry Management**: Organize skills, subagents, commands, MCPs, scripts, docs, and more
- **Multi-Target Support**: Install to different targets (Claude Code, Cursor, etc.)
- **Dependency Resolution**: Automatic topological sorting with cycle detection
- **Import External Files**: Scan and import existing markdown files
- **Git Integration**: Keep your registry synced with git
- **Multiple Output Formats**: Pretty terminal output, JSON for agents, quiet for piping

## Installation

<details>
<summary><b>macOS (Apple Silicon)</b></summary>

```bash
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_darwin_arm64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/
```
</details>

<details>
<summary><b>macOS (Intel)</b></summary>

```bash
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_darwin_amd64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/
```
</details>

<details>
<summary><b>Linux (amd64)</b></summary>

```bash
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_linux_amd64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/
```

Or via package manager:
```bash
# Debian/Ubuntu
curl -LO https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_linux_amd64.deb
sudo dpkg -i regis3_1.0.0_linux_amd64.deb

# RHEL/Fedora
curl -LO https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_linux_amd64.rpm
sudo rpm -i regis3_1.0.0_linux_amd64.rpm
```
</details>

<details>
<summary><b>Linux (arm64)</b></summary>

```bash
curl -L https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_linux_arm64.tar.gz | tar xz
sudo mv regis3 /usr/local/bin/
```

Or via package manager:
```bash
# Debian/Ubuntu
curl -LO https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_linux_arm64.deb
sudo dpkg -i regis3_1.0.0_linux_arm64.deb

# RHEL/Fedora
curl -LO https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_linux_arm64.rpm
sudo rpm -i regis3_1.0.0_linux_arm64.rpm
```
</details>

<details>
<summary><b>Windows (amd64)</b></summary>

1. Download [regis3_1.0.0_windows_amd64.zip](https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_windows_amd64.zip)
2. Extract `regis3.exe`
3. Move to a directory in your PATH (e.g., `C:\Program Files\regis3\`)
4. Add to PATH if needed

Or via PowerShell:
```powershell
Invoke-WebRequest -Uri "https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_windows_amd64.zip" -OutFile regis3.zip
Expand-Archive regis3.zip -DestinationPath .
Move-Item regis3.exe C:\Windows\System32\
```
</details>

<details>
<summary><b>Windows (arm64)</b></summary>

1. Download [regis3_1.0.0_windows_arm64.zip](https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_windows_arm64.zip)
2. Extract `regis3.exe`
3. Move to a directory in your PATH (e.g., `C:\Program Files\regis3\`)
4. Add to PATH if needed

Or via PowerShell:
```powershell
Invoke-WebRequest -Uri "https://github.com/okto-digital/regis3/releases/latest/download/regis3_1.0.0_windows_arm64.zip" -OutFile regis3.zip
Expand-Archive regis3.zip -DestinationPath .
Move-Item regis3.exe C:\Windows\System32\
```
</details>

<details>
<summary><b>Using Go</b></summary>

```bash
go install github.com/okto-digital/regis3/cmd/regis3@latest
```
</details>

<details>
<summary><b>From Source</b></summary>

```bash
git clone https://github.com/okto-digital/regis3.git
cd regis3
make install
```
</details>

<details>
<summary><b>Updating</b></summary>

Download the latest release for your platform and replace the binary, or:

```bash
# Using Go
go install github.com/okto-digital/regis3/cmd/regis3@latest

# From source
cd regis3 && git pull && make install
```
</details>

<details>
<summary><b>Uninstalling</b></summary>

Remove the binary:
```bash
sudo rm /usr/local/bin/regis3
# or if installed via Go
rm $(go env GOPATH)/bin/regis3
```

Remove regis3 data:
```bash
# Keep registry, remove config only
rm ~/.regis3/config.yaml

# Remove everything
rm -rf ~/.regis3
```
</details>

## Quick Start

```bash
# Initialize regis3 (first-time setup)
regis3 init

# List available items in registry
regis3 list

# Install a skill to current project
regis3 project add skill:git-conventions

# Show what's installed in current project
regis3 project status

# Update registry from git
regis3 update
```

## Configuration

regis3 uses a configuration file at `~/.regis3/config.yaml`:

```yaml
registry_path: ~/.regis3/registry
default_target: claude
output_format: pretty
```

### Configuration Commands

```bash
# Show current configuration
regis3 config

# Get a specific setting
regis3 config get registry

# Set a configuration value
regis3 config set registry ~/my-registry
```

## Registry Structure

```
~/.regis3/registry/
├── .build/
│   └── manifest.json      # Auto-generated index
├── import/                 # Staging area for imported files
├── skills/                 # Skill definitions
├── agents/                 # Subagent configurations
├── commands/               # Custom commands
├── stacks/                 # Preset combinations
└── philosophies/           # Coding philosophies
```

## Item Types

| Type | Description | Install Location |
|------|-------------|------------------|
| `skill` | Focused capability | `.claude/skills/{name}/` |
| `subagent` | Task-specific agent | `.claude/agents/` |
| `command` | Custom slash command | `.claude/commands/` |
| `mcp` | MCP server config | `.claude/mcp/` |
| `script` | Helper scripts | `.claude/scripts/` |
| `doc` | Documentation | `.claude/docs/` |
| `project` | Project templates | Merged into CLAUDE.md |
| `philosophy` | Coding philosophy | Merged into CLAUDE.md |
| `ruleset` | Coding rules | Merged into CLAUDE.md |
| `stack` | Item collection | Installs dependencies |
| `hook` | Event hooks | `.claude/hooks/` |
| `prompt` | Prompt templates | `.claude/prompts/` |

## Commands

### Registry Operations

```bash
# Build/rebuild the registry manifest
regis3 build

# Validate all items in the registry
regis3 validate

# Find orphaned files (not in manifest)
regis3 orphans

# Rebuild manifest after manual changes
regis3 reindex
```

### Discovery

```bash
# List all items
regis3 list

# List only skills
regis3 list --type skill

# List items with a specific tag
regis3 list --tag testing

# Search for items
regis3 search "git"

# Show details for an item
regis3 info skill:git-conventions
```

### Project Operations

```bash
# Install items to current project
regis3 project add skill:git-conventions

# Install multiple items
regis3 project add skill:git-conventions skill:testing

# Install a stack (all dependencies)
regis3 project add stack:base

# Preview installation (dry run)
regis3 project add skill:testing --dry-run

# Force reinstall
regis3 project add skill:testing --force
```

### Status & Updates

```bash
# Show installed items in current project
regis3 project status

# Update registry from git
regis3 update

# Remove items from current project
regis3 project remove skill:git-conventions
```

### Import External Files

```bash
# Scan a directory for markdown files
regis3 scan ~/Documents/prompts

# Preview what would be imported
regis3 scan ~/Documents/prompts --dry-run

# Process files in staging directory
regis3 import

# List pending files in staging
regis3 import --list
```

## Output Formats

regis3 supports three output formats:

```bash
# Pretty output (default) - colors and icons
regis3 list

# JSON output - for scripts and agents
regis3 list --format json

# Quiet output - minimal, one item per line
regis3 list --format quiet
```

## Creating Registry Items

Registry items are markdown files with YAML frontmatter:

```markdown
---
regis3:
  type: skill
  name: my-skill
  desc: Description of my skill
  tags: [tag1, tag2]
  deps:
    - skill:base-skill
---

# My Skill

Content of the skill...
```

### Required Fields

- `type`: One of the valid item types
- `name`: Unique name within the type
- `desc`: Short description

### Optional Fields

- `tags`: Array of tags for filtering
- `deps`: Array of dependencies (format: `type:name`)
- `files`: Additional files to include
- `status`: `stable`, `draft`, or `deprecated`
- `order`: Numeric order for merged items

## Shell Completions

Generate shell completions:

```bash
# Bash
regis3 completion bash > /etc/bash_completion.d/regis3

# Zsh
regis3 completion zsh > "${fpath[1]}/_regis3"

# Fish
regis3 completion fish > ~/.config/fish/completions/regis3.fish
```

## Development

```bash
# Build
make build

# Run tests
make test

# Run with coverage
make test-cover

# Format code
make fmt

# Lint
make lint

# Build for all platforms
make release
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `REGIS3_REGISTRY_PATH` | Override registry path |
| `REGIS3_DEFAULT_TARGET` | Override default target |
| `REGIS3_OUTPUT_FORMAT` | Override output format |
| `REGIS3_DEBUG` | Enable debug output |

## License

MIT License - see [LICENSE](LICENSE) for details.
