---
name: devcontainer-config
description: Manage devcontainer.json configuration for VS Code dev containers. Use when agent needs to inspect, modify, or prepare changes to dev container configuration including features, extensions, ports, and environment variables.
allowed-tools: Read, Write, Edit, Glob, Grep
---

# Dev Container Configuration Manager

**Version:** 1.0.0
**Purpose:** Inspect and prepare configuration changes for VS Code dev containers
**Scope:** Configuration management (non-destructive - user triggers rebuilds)

---

## Overview

This skill enables agents to manage their dev container environment configuration. It provides capabilities to:
- Inspect current `devcontainer.json` configuration
- Add/remove dev container features (languages, tools)
- Manage VS Code extensions
- Configure port forwarding
- Set environment variables
- Prepare rebuild instructions for the user

**IMPORTANT:** This skill prepares configuration changes. The user must trigger container rebuilds for changes to take effect.

---

## When to Use This Skill

Invoke this skill when:
- Agent needs to check its current dev container configuration
- Agent wants to add a new tool/language (feature)
- Agent needs to configure VS Code extensions
- Agent wants to set up port forwarding
- Environment variable configuration is needed
- User asks about dev container capabilities

---

## Key Concepts

### Dev Container Configuration File

Location: `.devcontainer/devcontainer.json`

```json
{
  "name": "Container Name",
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu",
  "features": {
    "ghcr.io/devcontainers/features/node:1": {},
    "ghcr.io/devcontainers/features/python:1": {}
  },
  "customizations": {
    "vscode": {
      "extensions": ["extension.id"]
    }
  },
  "forwardPorts": [3000, 8080],
  "postCreateCommand": "npm install",
  "remoteEnv": {
    "MY_VAR": "value"
  }
}
```

### Features

Modular units that add tools/languages to the container:

| Feature | Purpose |
|---------|---------|
| `ghcr.io/devcontainers/features/node:1` | Node.js, npm, yarn |
| `ghcr.io/devcontainers/features/python:1` | Python, pip |
| `ghcr.io/devcontainers/features/go:1` | Go language |
| `ghcr.io/devcontainers/features/rust:1` | Rust, cargo |
| `ghcr.io/devcontainers/features/docker-in-docker:1` | Docker inside container |
| `ghcr.io/devcontainers/features/github-cli:1` | GitHub CLI (gh) |
| `ghcr.io/devcontainers/features/azure-cli:1` | Azure CLI |
| `ghcr.io/devcontainers/features/aws-cli:1` | AWS CLI |
| `ghcr.io/devcontainers/features/kubectl-helm-minikube:1` | Kubernetes tools |
| `ghcr.io/devcontainers/features/git:1` | Latest Git |
| `ghcr.io/devcontainers/features/common-utils:2` | Common utilities |

---

## Operations

### 1. Inspect Configuration

```markdown
# Read current configuration
Read(".devcontainer/devcontainer.json")

# Find devcontainer.json in workspace
Glob("**/devcontainer.json")
```

### 2. Add Feature

Edit `devcontainer.json` to add a feature:

```json
"features": {
  "ghcr.io/devcontainers/features/node:1": {
    "version": "20"
  }
}
```

### 3. Add VS Code Extension

Edit `customizations.vscode.extensions`:

```json
"customizations": {
  "vscode": {
    "extensions": [
      "dbaeumer.vscode-eslint",
      "esbenp.prettier-vscode"
    ]
  }
}
```

### 4. Configure Port Forwarding

Edit `forwardPorts`:

```json
"forwardPorts": [3000, 8080, 5432]
```

### 5. Set Environment Variables

Edit `remoteEnv`:

```json
"remoteEnv": {
  "NODE_ENV": "development",
  "API_URL": "http://localhost:8080"
}
```

### 6. Add Post-Create Command

Edit `postCreateCommand`:

```json
"postCreateCommand": "npm install && npm run setup"
```

---

## Workflow

### Standard Configuration Change Workflow

**1. Locate Configuration**
```
Glob("**/devcontainer.json")
```

**2. Read Current State**
```
Read(".devcontainer/devcontainer.json")
```

**3. Prepare Changes**
- Use Edit tool to modify specific sections
- Preserve existing configuration
- Validate JSON structure

**4. Inform User**
After making changes, always inform the user:

```
Configuration updated in .devcontainer/devcontainer.json

Changes made:
- Added feature: ghcr.io/devcontainers/features/node:1
- Added extension: dbaeumer.vscode-eslint

To apply changes, rebuild the container:
- VS Code: Command Palette > "Dev Containers: Rebuild Container"
- CLI: devcontainer up --workspace-folder . --remove-existing-container
```

---

## Safety Guidelines

**DO:**
- Always read current configuration before modifying
- Preserve existing features and extensions (add to them)
- Validate JSON structure after changes
- Inform user about rebuild requirement
- Back up configuration if making major changes

**DO NOT:**
- Overwrite entire devcontainer.json (use Edit for specific changes)
- Remove existing features without user confirmation
- Assume changes take effect immediately
- Modify files outside .devcontainer directory

---

## Context7 Integration

**IMPORTANT:** For additional documentation on dev containers, VS Code, and Docker, use Context7 MCP:

```javascript
// Resolve library IDs
mcp__context7__resolve-library-id({ libraryName: "devcontainers" })
mcp__context7__resolve-library-id({ libraryName: "VS Code dev containers" })
mcp__context7__resolve-library-id({ libraryName: "Docker" })

// Fetch documentation
mcp__context7__get-library-docs({
  context7CompatibleLibraryID: "/devcontainers/devcontainers.github.io",
  topic: "features configuration"
})
```

**Available Context7 Libraries:**

| Topic | Library ID | Snippets |
|-------|------------|----------|
| Dev Containers Spec | `/devcontainers/devcontainers.github.io` | 202 |
| Dev Container Images | `/devcontainers/images` | 15,648 |
| Dev Container CLI | `/devcontainers/cli` | 117 |
| Dev Container Features | `/devcontainers/features` | 95 |
| Dev Container Templates | `/devcontainers/templates` | 215 |
| Docker Docs | `/websites/docs_docker_com` | 15,954 |

Use Context7 when:
- Looking up available features
- Understanding configuration options
- Troubleshooting container issues
- Finding best practices

---

## Example Use Cases

### Example 1: Add Node.js to Container

```
User: "I need Node.js in my dev container"

Agent:
1. Read(".devcontainer/devcontainer.json")
2. Edit to add feature:
   "ghcr.io/devcontainers/features/node:1": { "version": "20" }
3. Inform user to rebuild container
```

### Example 2: Check Current Configuration

```
User: "What features are in my dev container?"

Agent:
1. Glob("**/devcontainer.json")
2. Read the file
3. Report installed features, extensions, ports
```

### Example 3: Add Development Extensions

```
User: "Add ESLint and Prettier extensions"

Agent:
1. Read current configuration
2. Edit customizations.vscode.extensions array
3. Add: ["dbaeumer.vscode-eslint", "esbenp.prettier-vscode"]
4. Inform user to rebuild or reload
```

---

## Troubleshooting

### Cannot Find devcontainer.json

```
Glob("**/devcontainer.json")
Glob(".devcontainer/devcontainer.json")
```

If not found, offer to create one:
```json
{
  "name": "Development Container",
  "image": "mcr.microsoft.com/devcontainers/base:ubuntu",
  "features": {},
  "customizations": {
    "vscode": {
      "extensions": []
    }
  }
}
```

### Invalid JSON After Edit

- Read the file again
- Check for trailing commas
- Validate with JSON parser
- Fix syntax errors

### Feature Not Working

Suggest user:
1. Rebuild container completely
2. Check feature documentation via Context7
3. Verify feature compatibility with base image

---

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2024-12 | Initial release |

---

**Note:** This skill manages configuration only. Container lifecycle operations (start/stop/rebuild) require user action or the Dev Container CLI.
