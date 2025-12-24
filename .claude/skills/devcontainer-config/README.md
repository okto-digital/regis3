# Dev Container Configuration Manager

**Version:** 1.0.0
**Type:** Configuration Management Skill
**Scope:** VS Code Dev Containers

---

## Overview

This skill enables Claude Code agents to manage their VS Code dev container environment configuration. It provides a safe, non-destructive approach where the agent prepares configuration changes and the user triggers container rebuilds.

## Features

- Inspect current `devcontainer.json` configuration
- Add/remove dev container features (languages, tools, CLIs)
- Manage VS Code extensions
- Configure port forwarding
- Set environment variables
- Create new dev container configurations
- Generate rebuild instructions for users

## Installation

Copy the skill directory to your agent's skills folder:

```
.claude/skills/devcontainer-config/
├── SKILL.md      # Skill definition
├── handler.js    # Implementation
└── README.md     # This file
```

## Usage

### As Claude Code Skill

The skill is invoked automatically when relevant. The agent uses Read, Write, Edit, Glob, and Grep tools to manage configuration.

### Handler API

```javascript
const handler = require('./handler');

// Find devcontainer.json
const result = await handler.execute({ action: 'find' });

// Read configuration
const result = await handler.execute({ action: 'read' });

// Add a feature
const result = await handler.execute({
  action: 'add-feature',
  feature: 'node',  // Short name or full path
  featureOptions: { version: '20' }
});

// Add VS Code extension
const result = await handler.execute({
  action: 'add-extension',
  extension: 'dbaeumer.vscode-eslint'
});

// Add port forwarding
const result = await handler.execute({
  action: 'add-port',
  port: 3000
});

// Set environment variable
const result = await handler.execute({
  action: 'set-env',
  name: 'NODE_ENV',
  value: 'development'
});

// Create new configuration
const result = await handler.execute({
  action: 'create',
  name: 'My Dev Container',
  image: 'mcr.microsoft.com/devcontainers/base:ubuntu'
});

// List available features
const result = await handler.execute({ action: 'list-features' });
```

## Available Features (Short Names)

| Short Name | Full Feature Path |
|------------|-------------------|
| `node` | `ghcr.io/devcontainers/features/node:1` |
| `python` | `ghcr.io/devcontainers/features/python:1` |
| `go` | `ghcr.io/devcontainers/features/go:1` |
| `rust` | `ghcr.io/devcontainers/features/rust:1` |
| `java` | `ghcr.io/devcontainers/features/java:1` |
| `dotnet` | `ghcr.io/devcontainers/features/dotnet:2` |
| `docker-in-docker` | `ghcr.io/devcontainers/features/docker-in-docker:2` |
| `github-cli` | `ghcr.io/devcontainers/features/github-cli:1` |
| `azure-cli` | `ghcr.io/devcontainers/features/azure-cli:1` |
| `aws-cli` | `ghcr.io/devcontainers/features/aws-cli:1` |
| `kubectl-helm-minikube` | `ghcr.io/devcontainers/features/kubectl-helm-minikube:1` |
| `terraform` | `ghcr.io/devcontainers/features/terraform:1` |
| `git` | `ghcr.io/devcontainers/features/git:1` |
| `common-utils` | `ghcr.io/devcontainers/features/common-utils:2` |

## Context7 Integration

For additional documentation, use Context7 MCP to look up:

| Topic | Library ID |
|-------|------------|
| Dev Containers Spec | `/devcontainers/devcontainers.github.io` |
| Dev Container Images | `/devcontainers/images` |
| Dev Container CLI | `/devcontainers/cli` |
| Dev Container Features | `/devcontainers/features` |
| Dev Container Templates | `/devcontainers/templates` |
| Docker Documentation | `/websites/docs_docker_com` |
| VS Code Dev Containers | `/websites/code_visualstudio_devcontainers_containers` |

Example:
```javascript
// Resolve library
mcp__context7__resolve-library-id({ libraryName: "devcontainers" })

// Get documentation
mcp__context7__get-library-docs({
  context7CompatibleLibraryID: "/devcontainers/features",
  topic: "python configuration"
})
```

## Safety Model

This skill follows a **prepare-and-inform** pattern:

1. **Agent prepares** - Modifies `devcontainer.json`
2. **Agent informs** - Tells user what changed and how to apply
3. **User triggers** - User rebuilds container to apply changes

This prevents the agent from accidentally breaking its own environment.

## Example Conversations

### Adding a Tool

```
User: I need Python in my dev container

Agent: I'll add Python to your dev container configuration.

[Reads .devcontainer/devcontainer.json]
[Adds ghcr.io/devcontainers/features/python:1]

Done! I've added Python to your dev container configuration.

To apply this change, rebuild the container:
- VS Code: Cmd/Ctrl+Shift+P > "Dev Containers: Rebuild Container"
```

### Checking Configuration

```
User: What's installed in my dev container?

Agent: Let me check your dev container configuration.

[Reads .devcontainer/devcontainer.json]

Your dev container includes:
- Base image: mcr.microsoft.com/devcontainers/base:ubuntu
- Features: node:1, python:1, github-cli:1
- Extensions: eslint, prettier
- Forwarded ports: 3000, 8080
```

## Troubleshooting

### devcontainer.json Not Found

The skill searches these locations:
1. `.devcontainer/devcontainer.json`
2. `.devcontainer.json`
3. `devcontainer.json`

If not found, use the `create` action to generate one.

### JSON Parse Errors

The handler supports JSONC (JSON with comments). If you encounter parse errors:
1. Check for trailing commas
2. Verify bracket matching
3. Use a JSON validator

### Changes Not Taking Effect

Remember: configuration changes require a container rebuild. The agent will provide instructions after each change.

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-12 | Initial release |

## License

Part of the Claude Code Ecosystem.
